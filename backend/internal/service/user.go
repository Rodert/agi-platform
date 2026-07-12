package service

import (
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo   *repository.UserRepository
	creditRepo *repository.CreditRepository
}

func NewUserService(userRepo *repository.UserRepository, creditRepo *repository.CreditRepository) *UserService {
	return &UserService{
		userRepo:   userRepo,
		creditRepo: creditRepo,
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
	}, nil
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
