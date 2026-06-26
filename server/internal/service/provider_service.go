package service

import (
	"context"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

type ProviderService interface {
	ListEnabled(ctx context.Context) ([]model.Provider, error)
}

type providerService struct {
	providers repository.ProviderRepository
}

func NewProviderService(providers repository.ProviderRepository) ProviderService {
	return &providerService{providers: providers}
}

func (s *providerService) ListEnabled(ctx context.Context) ([]model.Provider, error) {
	return s.providers.ListEnabled(ctx)
}
