package service

import (
	"fmt"
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
		jwtConfig:     jwtConfig,
		db:            db,
	}
}

// Register 用户注册
func (s *AuthService) Register(req *dto.RegisterRequest) (*dto.AuthResponse, error) {
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

	// 3. 验证验证码
	if err := s.verifyCode(req.Email, req.Code, "register"); err != nil {
		return nil, err
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

	// 7. 开启事务
	var user *model.User
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		user = &model.User{
			Email:        req.Email,
			PasswordHash: passwordHash,
			Name:         req.Email[:4] + "****", // 默认昵称
			Level:        "free",
			InviteCode:   inviteCode,
			InvitedBy:    inviterID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// 创建积分账户并赠送新用户积分
		newUserGift := 120 // 从配置读取
		if err := s.creditService.CreateAccountWithTx(tx, user.ID, newUserGift); err != nil {
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

	// 8. 生成 Token
	token, err := jwt.GenerateToken(user.ID, user.Email, s.jwtConfig)
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

// Login 用户登录
func (s *AuthService) Login(req *dto.LoginRequest) (*dto.AuthResponse, error) {
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
	token, err := jwt.GenerateToken(user.ID, user.Email, s.jwtConfig)
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
	if req.Type == "login" && !exists {
		return errors.ErrUserNotFound
	}

	// 2. 生成验证码
	code := utils.GenerateCode()

	// 3. 保存验证码
	verifyCode := &model.VerificationCode{
		Email:     req.Email,
		Code:      code,
		Type:      req.Type,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		CreatedAt: time.Now(),
	}
	if err := s.codeRepo.Create(verifyCode); err != nil {
		return err
	}

	// 4. 发送邮件
	if err := s.emailService.SendVerificationCode(req.Email, code); err != nil {
		return err
	}

	return nil
}

// verifyCode 验证验证码
func (s *AuthService) verifyCode(email, code, codeType string) error {
	// 1. 查找最新的验证码
	verifyCode, err := s.codeRepo.FindLatest(email, codeType)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
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
		return errors.ErrCodeExpired
	}

	// 4. 标记为已使用
	if err := s.codeRepo.MarkAsUsed(verifyCode.ID); err != nil {
		return err
	}

	return nil
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
