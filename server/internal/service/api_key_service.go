package service

import (
	"context"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

type APIKeyService interface {
	Create(ctx context.Context, userID uint64, name string) (*CreateAPIKeyResult, error)
	List(ctx context.Context, userID uint64) ([]model.APIKey, error)
	Revoke(ctx context.Context, userID uint64, id uint64) error
	Authenticate(ctx context.Context, plain string) (*model.APIKey, error)
}

type CreateAPIKeyResult struct {
	APIKey model.APIKey `json:"api_key"`
	Plain  string       `json:"plain"`
}

type apiKeyService struct {
	repos repository.Repositories
	auth  auth.Manager
}

func NewAPIKeyService(repos repository.Repositories, authManager auth.Manager) APIKeyService {
	return &apiKeyService{
		repos: repos,
		auth:  authManager,
	}
}

func (s *apiKeyService) Create(ctx context.Context, userID uint64, name string) (*CreateAPIKeyResult, error) {
	plain, prefix, hash, err := s.auth.NewAPIKey()
	if err != nil {
		return nil, err
	}

	key := &model.APIKey{
		UserID:    userID,
		Name:      name,
		KeyPrefix: prefix,
		KeyHash:   hash,
		Status:    "active",
	}
	if err := s.repos.APIKeys.Create(ctx, nil, key); err != nil {
		return nil, err
	}

	return &CreateAPIKeyResult{
		APIKey: *key,
		Plain:  plain,
	}, nil
}

func (s *apiKeyService) List(ctx context.Context, userID uint64) ([]model.APIKey, error) {
	return s.repos.APIKeys.ListByUserID(ctx, userID)
}

func (s *apiKeyService) Revoke(ctx context.Context, userID uint64, id uint64) error {
	return s.repos.APIKeys.Revoke(ctx, userID, id)
}

func (s *apiKeyService) Authenticate(ctx context.Context, plain string) (*model.APIKey, error) {
	hash := s.auth.HashAPIKey(plain)
	key, err := s.repos.APIKeys.FindActiveByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	_ = s.repos.APIKeys.TouchLastUsed(ctx, key.ID)
	return key, nil
}
