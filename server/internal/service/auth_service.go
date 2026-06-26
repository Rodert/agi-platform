package service

import (
	"context"
	"errors"
	"strings"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

type AuthService interface {
	Register(ctx context.Context, req RegisterRequest) (*AuthResult, error)
	Login(ctx context.Context, req LoginRequest) (*AuthResult, error)
	Me(ctx context.Context, userID uint64) (*model.User, error)
}

type RegisterRequest struct {
	Email    string
	Password string
	Nickname string
}

type LoginRequest struct {
	Email    string
	Password string
}

type AuthResult struct {
	User  model.User  `json:"user"`
	Token *auth.Token `json:"token"`
}

type authService struct {
	repos               repository.Repositories
	auth                auth.Manager
	registerGiftCredits int64
}

func NewAuthService(repos repository.Repositories, authManager auth.Manager, registerGiftCredits int64) AuthService {
	return &authService{
		repos:               repos,
		auth:                authManager,
		registerGiftCredits: registerGiftCredits,
	}
}

func (s *authService) Register(ctx context.Context, req RegisterRequest) (*AuthResult, error) {
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
		PasswordHash: passwordHash,
		Nickname:     defaultNickname(req.Nickname, email),
		Credits:      s.registerGiftCredits,
		Status:       "active",
	}

	if err := s.repos.Tx.Transaction(ctx, func(tx repository.Tx) error {
		if err := s.repos.Users.Create(ctx, tx, user); err != nil {
			return err
		}
		if s.registerGiftCredits <= 0 {
			return nil
		}
		relatedID := user.ID
		return s.repos.Wallets.CreateLog(ctx, tx, &model.WalletLog{
			UserID:        user.ID,
			Type:          "register_gift",
			Amount:        s.registerGiftCredits,
			BalanceBefore: 0,
			BalanceAfter:  s.registerGiftCredits,
			RelatedType:   "user",
			RelatedID:     &relatedID,
			Remark:        "register gift credits",
			OperatorType:  "system",
		})
	}); err != nil {
		return nil, err
	}

	token, err := s.auth.IssueUserToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: *user, Token: token}, nil
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	user, err := s.repos.Users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if user.Status != "active" || !s.auth.CheckPassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.auth.IssueUserToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{User: *user, Token: token}, nil
}

func (s *authService) Me(ctx context.Context, userID uint64) (*model.User, error) {
	return s.repos.Users.FindByID(ctx, userID)
}

func defaultNickname(nickname string, email string) string {
	nickname = strings.TrimSpace(nickname)
	if nickname != "" {
		return nickname
	}
	parts := strings.Split(email, "@")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "User"
}
