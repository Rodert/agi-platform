package repository

import (
	"context"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type WalletRepository interface {
	CreateLog(ctx context.Context, tx Tx, log *model.WalletLog) error
}

type GormWalletRepository struct {
	db *gorm.DB
}

func NewGormWalletRepository(db *gorm.DB) *GormWalletRepository {
	return &GormWalletRepository{db: db}
}

func (r *GormWalletRepository) CreateLog(ctx context.Context, tx Tx, log *model.WalletLog) error {
	db := r.db
	if txDB := dbFromTx(tx); txDB != nil {
		db = txDB
	}
	return db.WithContext(ctx).Create(log).Error
}
