package service

import (
	"context"

	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

type WalletService interface {
	List(ctx context.Context, limit int, offset int) ([]model.WalletLog, error)
	ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.WalletLog, error)
}

type walletService struct {
	repos repository.Repositories
}

func NewWalletService(repos repository.Repositories) WalletService {
	return &walletService{repos: repos}
}

func (s *walletService) List(ctx context.Context, limit int, offset int) ([]model.WalletLog, error) {
	limit, offset = normalizePage(limit, offset)
	return s.repos.Wallets.List(ctx, limit, offset)
}

func (s *walletService) ListForUser(ctx context.Context, userID uint64, limit int, offset int) ([]model.WalletLog, error) {
	if userID == 0 {
		return nil, ErrInvalidRequest
	}
	limit, offset = normalizePage(limit, offset)
	return s.repos.Wallets.ListByUserID(ctx, userID, limit, offset)
}
