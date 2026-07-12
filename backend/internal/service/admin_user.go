package service

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

// GetUserList 获取用户列表。
func (s *AdminService) GetUserList(req *dto.UserListRequest) (*dto.UserListResponse, error) {
	users, total, err := s.userRepo.GetList(req.Page, req.PageSize, req.Username, req.Email)
	if err != nil {
		return nil, errors.NewWithDetails(errors.ErrCodeInternalServer, "查询用户列表失败", err.Error())
	}
	return &dto.UserListResponse{List: users, Total: total}, nil
}

// CreateUser 创建用户。
func (s *AdminService) CreateUser(req *dto.CreateUserRequest) error {
	exists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return err
	}
	if exists {
		return errors.ErrUserExists
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return err
	}
	var inviteCode string
	for i := 0; i < 10; i++ {
		candidate := utils.GenerateInviteCode()
		exists, err := s.userRepo.ExistsByInviteCode(candidate)
		if err != nil {
			return err
		}
		if !exists {
			inviteCode = candidate
			break
		}
	}
	if inviteCode == "" {
		return errors.New(errors.ErrCodeInternalServer, "生成邀请码失败")
	}
	now := time.Now()
	return s.userRepo.Create(&model.User{
		Email: req.Email, PasswordHash: hashedPassword, Name: req.Username,
		Level: "free", InviteCode: inviteCode, CreatedAt: now, UpdatedAt: now,
	})
}

func (s *AdminService) UpdateUser(userID int64, req *dto.AdminUpdateUserRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return errors.ErrUserNotFound }
		return err
	}
	if req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmailExceptID(req.Email, userID)
		if err != nil { return err }
		if exists { return errors.ErrUserExists }
	}
	user.Name = req.Username
	user.Email = req.Email
	user.Level = req.Level
	if req.Password != "" {
		hash, err := utils.HashPassword(req.Password)
		if err != nil { return err }
		user.PasswordHash = hash
	}
	user.UpdatedAt = time.Now()
	return s.userRepo.Update(user)
}

// UpdateUserStatus 当前用户表没有启停字段。
func (s *AdminService) UpdateUserStatus(userID int64, _ bool) error {
	if _, err := s.userRepo.FindByID(userID); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrUserNotFound
		}
		return err
	}
	return errors.New(errors.ErrCodeBadRequest, "当前用户模型暂不支持启停状态")
}
