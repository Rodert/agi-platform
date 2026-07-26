package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	codeRepo       *repository.VerificationCodeRepository
	configRepo     *repository.ConfigRepository
	creditService  *CreditService
	inviteService  *InvitationService
	emailService   *EmailService
	sessionRepo    *repository.UserSessionRepository
	adminRepo      *repository.AdminRepository
	jwtConfig      *config.JWTConfig
	db             *gorm.DB
}

func NewAuthService(
	userRepo *repository.UserRepository,
	codeRepo *repository.VerificationCodeRepository,
	configRepo *repository.ConfigRepository,
	creditService *CreditService,
	inviteService *InvitationService,
	emailService *EmailService,
	sessionRepo *repository.UserSessionRepository,
	adminRepo *repository.AdminRepository,
	jwtConfig *config.JWTConfig,
	db *gorm.DB,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		codeRepo:      codeRepo,
		configRepo:    configRepo,
		creditService: creditService,
		inviteService: inviteService,
		emailService:  emailService,
		sessionRepo: sessionRepo,
		adminRepo:   adminRepo,
		jwtConfig:     jwtConfig,
		db:            db,
	}
}

// Register 用户注册
func (s *AuthService) Register(req *dto.RegisterRequest, device, ip string) (*dto.AuthResponse, error) {
	// 1. 验证密码一致性
	if req.Password != req.ConfirmPassword {
		return nil, errors.New(errors.ErrCodeBadRequest, "两次输入的密码不一致")
	}

	// 2. 检查邮箱是否已存在
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.ErrUserExists
	}

	// 3. 根据用户默认设置决定是否要求邮箱验证码。
	requireVerification, err := s.registerEmailVerificationRequired()
	if err != nil {
		return nil, err
	}
	if requireVerification {
		if req.Code == "" {
			return nil, errors.New(errors.ErrCodeBadRequest, "请输入邮箱验证码")
		}
		if err := s.verifyCode(req.Email, req.Code, "register"); err != nil {
			return nil, err
		}
	}

	// 4. 验证邀请码（如果有）
	var inviterID int64
	if req.InviteCode != "" {
		inviter, err := s.userRepo.FindByInviteCode(req.InviteCode)
		if err != nil {
			return nil, errors.ErrInvalidInviteCode
		}
		inviterID = inviter.ID
	}

	// 5. 生成唯一邀请码
	inviteCode, err := s.generateUniqueInviteCode()
	if err != nil {
		return nil, err
	}

	// 6. 加密密码
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 7. 读取新用户默认设置。缺省时赠送 5 灵感值。
	defaults, err := s.userRegistrationDefaults()
	if err != nil {
		return nil, err
	}

	// 8. 开启事务
	var user *model.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		user = &model.User{
			Email:        req.Email,
			PasswordHash: passwordHash,
			Name:         req.Email[:4] + "****", // 默认昵称
			Avatar:       defaults.avatar,
			Level:        defaults.level,
			InviteCode:   inviteCode,
			InvitedBy:    inviterID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 创建积分账户并赠送新用户积分
		if err := s.creditService.CreateAccountWithTx(tx, user.ID, defaults.giftAmount); err != nil {
			return err
		}

		// 处理邀请奖励
		if inviterID > 0 {
			if err := s.inviteService.ProcessRegisterRewardWithTx(tx, inviterID, user.ID, inviteCode); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 9. 生成 Token
	token, err := s.createSessionToken(user, device, ip)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: &dto.UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Name:       user.Name,
			Avatar:     user.Avatar,
			Level:      user.Level,
			InviteCode: user.InviteCode,
		},
	}, nil
}

type userRegistrationDefaults struct { giftAmount int; level, avatar string }

func (s *AuthService) userRegistrationDefaults() (*userRegistrationDefaults, error) {
	amountValue, err := s.configRepo.GetSystemConfigValue("new_user_gift_amount", "5")
	if err != nil { return nil, err }
	amount, err := strconv.Atoi(amountValue)
	if err != nil || amount < 0 {
		return nil, fmt.Errorf("invalid new_user_gift_amount configuration")
	}
	level, err := s.configRepo.GetSystemConfigValue("default_user_level", "free")
	if err != nil { return nil, err }
	if level != "free" && level != "member" && level != "pro" { return nil, fmt.Errorf("invalid default_user_level configuration") }
	avatar, err := s.configRepo.GetSystemConfigValue("default_user_avatar", "")
	if err != nil { return nil, err }
	return &userRegistrationDefaults{giftAmount: amount, level: level, avatar: avatar}, nil
}

func (s *AuthService) registerEmailVerificationRequired() (bool, error) {
	value, err := s.configRepo.GetSystemConfigValue("register_email_verification", "true")
	if err != nil { return false, err }
	return value == "true", nil
}

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest, device, ip string) (*dto.AuthResponse, error) {
	// 1. 查找用户
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	// 2. 验证密码或验证码
	if req.Type == "password" {
		if req.Password == "" {
			return nil, errors.New(errors.ErrCodeBadRequest, "请输入密码")
		}
		if !utils.CheckPassword(req.Password, user.PasswordHash) {
			return nil, errors.ErrInvalidPassword
		}
	} else if req.Type == "code" {
		if req.Code == "" {
			return nil, errors.New(errors.ErrCodeBadRequest, "请输入验证码")
		}
		if err := s.verifyCode(req.Email, req.Code, "login"); err != nil {
			return nil, err
		}
	}

	// 3. 生成 Token
	token, err := s.createSessionToken(user, device, ip)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: &dto.UserInfo{
			ID:         user.ID,
			Email:      user.Email,
			Name:       user.Name,
			Avatar:     user.Avatar,
			Level:      user.Level,
			InviteCode: user.InviteCode,
		},
	}, nil
}

func (s *AuthService) createSessionToken(user *model.User, device, ip string) (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil { return "", err }
	sessionID := hex.EncodeToString(bytes)
	now := time.Now()
	session := &model.UserSession{ID: sessionID, UserID: user.ID, Device: device, IP: ip, CreatedAt: now, UpdatedAt: now, ExpiresAt: now.Add(s.jwtConfig.GetExpiration())}
	if err := s.sessionRepo.Create(session); err != nil { return "", err }
	return jwt.GenerateToken(user.ID, user.Email, sessionID, s.jwtConfig)
}

// SendCode 发送验证码
func (s *AuthService) SendCode(req *dto.SendCodeRequest) error {
	// 1. 检查邮箱是否存在（注册时不应存在，登录时应存在）
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return err
	}

	if req.Type == "register" && exists {
		return errors.ErrUserExists
	}
	if (req.Type == "login" || req.Type == "reset") && !exists {
		return errors.ErrUserNotFound
	}

	// 2. Re-send a still-valid code so a delayed earlier email does not become unusable.
	verifyCode, err := s.codeRepo.FindActive(req.Email, req.Type)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == gorm.ErrRecordNotFound {
		verifyCode = &model.VerificationCode{
			Email:     req.Email,
			Code:      utils.GenerateCode(),
			Type:      req.Type,
			ExpiresAt: time.Now().Add(5 * time.Minute),
			CreatedAt: time.Now(),
		}
		if err := s.codeRepo.Create(verifyCode); err != nil {
			return err
		}
	}

	// 3. Send the email and retain the code in the administrator-only audit log.
	if err := s.emailService.SendVerificationCode(req.Email, verifyCode.Code); err != nil {
		s.recordVerificationEvent("send_verification_code_failed", req.Email, verifyCode.Code, req.Type, "邮箱验证码发送失败")
		return err
	}
	s.recordVerificationEvent("send_verification_code", req.Email, verifyCode.Code, req.Type, "邮箱验证码已发送")

	return nil
}

// verifyCode 验证验证码
func (s *AuthService) verifyCode(email, code, codeType string) error {
	// 1. Match the submitted code among all usable codes. This tolerates email
	// delivery ordering while every code remains valid for only five minutes.
	verifyCode, err := s.codeRepo.FindMatchingActive(email, code, codeType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.recordVerificationEvent("verify_verification_code_failed", email, code, codeType, "邮箱验证码校验失败")
			return errors.ErrInvalidCode
		}
		return err
	}

	// 2. 检查验证码是否正确
	if verifyCode.Code != code {
		return errors.ErrInvalidCode
	}

	// 3. 检查是否过期
	if time.Now().After(verifyCode.ExpiresAt) {
		s.recordVerificationEvent("verify_verification_code_failed", email, code, codeType, "邮箱验证码已过期")
		return errors.ErrCodeExpired
	}

	// 4. 标记为已使用
	if err := s.codeRepo.MarkAsUsed(verifyCode.ID); err != nil {
		return err
	}
	s.recordVerificationEvent("verify_verification_code", email, code, codeType, "邮箱验证码校验成功")

	return nil
}

// recordVerificationEvent keeps public-account events visible to administrators.
// AdminID 0 denotes a system-triggered event and has no foreign-key dependency.
func (s *AuthService) recordVerificationEvent(action, email, code, codeType, description string) {
	if s.adminRepo == nil {
		return
	}
	_ = s.adminRepo.CreateLog(&model.AdminLog{
		AdminID:     0,
		Action:      action,
		TargetType:  "email_verification",
		Description: description,
		AfterData:   fmt.Sprintf(`{"email":%q,"code":%q,"type":%q}`, email, code, codeType),
		CreatedAt:   time.Now(),
	})
}

// generateUniqueInviteCode 生成唯一的邀请码
func (s *AuthService) generateUniqueInviteCode() (string, error) {
	for i := 0; i < 10; i++ {
		code := utils.GenerateInviteCode()
		exists, err := s.userRepo.ExistsByInviteCode(code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", fmt.Errorf("生成邀请码失败")
}
