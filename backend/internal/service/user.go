package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo   *repository.UserRepository
	creditRepo *repository.CreditRepository
	codeRepo *repository.VerificationCodeRepository
	sessionRepo *repository.UserSessionRepository
}

func NewUserService(userRepo *repository.UserRepository, creditRepo *repository.CreditRepository, codeRepo *repository.VerificationCodeRepository, sessionRepo *repository.UserSessionRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		creditRepo: creditRepo,
		codeRepo: codeRepo, sessionRepo: sessionRepo,
	}
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(userID int64) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}
	account, err := s.creditRepo.GetOrCreateAccount(userID)
	if err != nil {
		return nil, err
	}

	return &dto.UserProfileResponse{
		ID:         user.ID,
		Email:      user.Email,
		Name:       user.Name,
		Avatar:     user.Avatar,
		Bio:        user.Bio,
		Level:      user.Level,
		Balance:    account.Balance,
		InviteCode: user.InviteCode,
		Following:  0, // TODO: 从关注表统计
		Followers:  0, // TODO: 从关注表统计
		CreatedAt:  user.CreatedAt.Format("2006-01-02 15:04:05"),
		Phone: phoneValue(user.Phone),
	}, nil
}

func (s *UserService) BindPhone(userID int64, req *dto.BindPhoneRequest) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil { return nil, err }
	code, err := s.codeRepo.FindLatest(user.Email, "reset")
	if err != nil || code.Code != req.Code || time.Now().After(code.ExpiresAt) { return nil, errors.ErrInvalidCode }
	if err := s.codeRepo.MarkAsUsed(code.ID); err != nil { return nil, err }
	exists, err := s.userRepo.ExistsByPhoneExceptID(req.Phone, userID)
	if err != nil { return nil, err }; if exists { return nil, errors.New(errors.ErrCodeBadRequest, "该手机号已被绑定") }
	user.Phone = &req.Phone
	if err := s.userRepo.Update(user); err != nil { return nil, err }
	return s.GetProfile(userID)
}

func (s *UserService) ChangePassword(userID int64, currentSession string, req *dto.ChangePasswordRequest) error {
	user, err := s.userRepo.FindByID(userID); if err != nil { return err }
	if !utils.CheckPassword(req.CurrentPassword, user.PasswordHash) { return errors.ErrInvalidPassword }
	hash, err := utils.HashPassword(req.NewPassword); if err != nil { return err }
	user.PasswordHash = hash
	if err := s.userRepo.Update(user); err != nil { return err }
	return s.sessionRepo.RevokeOther(userID, currentSession)
}

func (s *UserService) ListSessions(userID int64, currentSession string) ([]*dto.UserSessionResponse, error) {
	sessions, err := s.sessionRepo.ListActive(userID); if err != nil { return nil, err }
	result := make([]*dto.UserSessionResponse, 0, len(sessions))
	for _, item := range sessions { result = append(result, &dto.UserSessionResponse{ID:item.ID, Device:item.Device, IP:item.IP, CreatedAt:item.CreatedAt.Format("2006-01-02 15:04:05"), Current:item.ID == currentSession}) }
	return result, nil
}

func (s *UserService) RevokeSession(userID int64, sessionID string) error { return s.sessionRepo.Revoke(userID, sessionID) }

func phoneValue(phone *string) string {
	if phone == nil {
		return ""
	}
	return *phone
}

// GetCreditLedgers returns only the authenticated user's credit history.
func (s *UserService) GetCreditLedgers(userID int64, page, pageSize int) ([]*dto.CreditLedgerResponse, int64, error) {
	ledgers, total, err := s.creditRepo.GetLedgers(userID, "", "", nil, nil, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*dto.CreditLedgerResponse, 0, len(ledgers))
	for _, ledger := range ledgers {
		items = append(items, &dto.CreditLedgerResponse{
			ID: ledger.ID, Type: ledger.Type, Amount: ledger.Amount, Title: ledger.Title,
			SourceType: ledger.SourceType, BalanceAfter: ledger.BalanceAfter,
			CreatedAt: ledger.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return items, total, nil
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID int64, req *dto.UpdateUserRequest) (*dto.UserProfileResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}
	return s.GetProfile(userID)
}
