package service

import (
	"context"
	"errors"
	"strings"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

type AdminService interface {
	Login(ctx context.Context, req AdminLoginRequest) (*AdminAuthResult, error)
	Me(ctx context.Context, adminID uint64) (*model.AdminUser, error)
	ChangePassword(ctx context.Context, adminID uint64, req ChangePasswordRequest) error
	ListUsers(ctx context.Context, limit int, offset int) ([]model.User, error)
	CreateUser(ctx context.Context, req AdminSaveUserRequest) (*model.User, error)
	UpdateUser(ctx context.Context, req AdminSaveUserRequest) (*model.User, error)
	AdjustUserCredits(ctx context.Context, req AdjustUserCreditsRequest) (*model.User, error)
}

type AdminLoginRequest struct {
	Username string
	Password string
}

type AdminAuthResult struct {
	Admin model.AdminUser `json:"admin"`
	Token *auth.Token     `json:"token"`
}

type AdminSaveUserRequest struct {
	UserID    uint64
	Email     string
	Phone     string
	Password  string
	Nickname  string
	AvatarURL string
	Status    string
}

type AdjustUserCreditsRequest struct {
	AdminID uint64
	UserID  uint64
	Amount  int64
	Remark  string
}

type adminService struct {
	repos repository.Repositories
	auth  auth.Manager
}

func NewAdminService(repos repository.Repositories, authManager auth.Manager) AdminService {
	return &adminService{
		repos: repos,
		auth:  authManager,
	}
}

func (s *adminService) Login(ctx context.Context, req AdminLoginRequest) (*AdminAuthResult, error) {
	username := strings.TrimSpace(req.Username)
	if username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}

	admin, err := s.repos.Admins.FindByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if admin.Status != "active" || !s.auth.CheckPassword(admin.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.IssueAdminToken(admin.ID)
	if err != nil {
		return nil, err
	}
	_ = s.repos.Admins.TouchLastLogin(ctx, admin.ID)

	return &AdminAuthResult{Admin: *admin, Token: token}, nil
}

func (s *adminService) Me(ctx context.Context, adminID uint64) (*model.AdminUser, error) {
	return s.repos.Admins.FindByID(ctx, adminID)
}

func (s *adminService) ChangePassword(ctx context.Context, adminID uint64, req ChangePasswordRequest) error {
	if adminID == 0 || len(req.NewPassword) < 6 || strings.TrimSpace(req.CurrentPassword) == "" {
		return ErrInvalidRequest
	}
	admin, err := s.repos.Admins.FindByID(ctx, adminID)
	if err != nil {
		return err
	}
	if !s.auth.CheckPassword(admin.PasswordHash, req.CurrentPassword) {
		return ErrInvalidCredentials
	}
	passwordHash, err := s.auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	return s.repos.Admins.UpdatePassword(ctx, adminID, passwordHash)
}

func (s *adminService) ListUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Users.List(ctx, limit, offset)
}

func (s *adminService) CreateUser(ctx context.Context, req AdminSaveUserRequest) (*model.User, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || len(req.Password) < 6 {
		return nil, ErrInvalidRequest
	}
	if _, err := s.repos.Users.FindByEmail(ctx, email); err == nil {
		return nil, ErrEmailAlreadyExists
	} else if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	passwordHash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Email:        &email,
		Phone:        optionalString(req.Phone),
		PasswordHash: passwordHash,
		Nickname:     defaultNickname(req.Nickname, email),
		AvatarURL:    strings.TrimSpace(req.AvatarURL),
		Status:       normalizeUserStatus(req.Status),
	}
	if err := s.repos.Users.Create(ctx, nil, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *adminService) UpdateUser(ctx context.Context, req AdminSaveUserRequest) (*model.User, error) {
	if req.UserID == 0 {
		return nil, ErrInvalidRequest
	}

	user, err := s.repos.Users.FindByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" {
		return nil, ErrInvalidRequest
	}
	if existing, err := s.repos.Users.FindByEmail(ctx, email); err == nil && existing.ID != req.UserID {
		return nil, ErrEmailAlreadyExists
	} else if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	user.Email = &email
	user.Phone = optionalString(req.Phone)
	user.Nickname = defaultNickname(req.Nickname, email)
	user.AvatarURL = strings.TrimSpace(req.AvatarURL)
	user.Status = normalizeUserStatus(req.Status)
	if strings.TrimSpace(req.Password) != "" {
		if len(req.Password) < 6 {
			return nil, ErrInvalidRequest
		}
		passwordHash, err := s.auth.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = passwordHash
	}

	if err := s.repos.Users.Update(ctx, nil, user); err != nil {
		return nil, err
	}
	return s.repos.Users.FindByID(ctx, req.UserID)
}

func (s *adminService) AdjustUserCredits(ctx context.Context, req AdjustUserCreditsRequest) (*model.User, error) {
	if req.AdminID == 0 || req.UserID == 0 || req.Amount == 0 {
		return nil, ErrInvalidRequest
	}

	var updated *model.User
	err := s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		before, err := s.repos.Users.FindByID(ctx, req.UserID)
		if err != nil {
			return err
		}

		if req.Amount > 0 {
			updated, err = s.repos.Users.AddCredits(ctx, tx, req.UserID, req.Amount)
		} else {
			updated, err = s.repos.Users.DeductCredits(ctx, tx, req.UserID, -req.Amount)
		}
		if err != nil {
			return err
		}

		relatedID := req.UserID
		logType := "admin_add"
		if req.Amount < 0 {
			logType = "admin_deduct"
		}
		adminID := req.AdminID
		return s.repos.Wallets.CreateLog(ctx, tx, &model.WalletLog{
			UserID:        req.UserID,
			Type:          logType,
			Amount:        req.Amount,
			BalanceBefore: before.Credits,
			BalanceAfter:  updated.Credits,
			RelatedType:   "user",
			RelatedID:     &relatedID,
			Remark:        req.Remark,
			OperatorType:  "admin",
			OperatorID:    &adminID,
		})
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) && req.Amount < 0 {
			return nil, ErrInsufficientCredits
		}
		return nil, err
	}
	return updated, nil
}

func normalizePage(limit int, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func optionalString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func normalizeUserStatus(status string) string {
	switch strings.TrimSpace(status) {
	case "disabled":
		return "disabled"
	default:
		return "active"
	}
}
