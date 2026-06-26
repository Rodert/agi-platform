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
	ListUsers(ctx context.Context, limit int, offset int) ([]model.User, error)
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

func (s *adminService) ListUsers(ctx context.Context, limit int, offset int) ([]model.User, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Users.List(ctx, limit, offset)
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
