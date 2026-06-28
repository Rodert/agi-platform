package repository

import (
	"context"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type WalletRepository interface {
	CreateLog(ctx context.Context, tx Tx, log *model.WalletLog) error
	List(ctx context.Context, limit int, offset int) ([]model.WalletLog, error)
	ListByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.WalletLog, error)
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

func (r *GormWalletRepository) List(ctx context.Context, limit int, offset int) ([]model.WalletLog, error) {
	var logs []model.WalletLog
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}

func (r *GormWalletRepository) ListByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.WalletLog, error) {
	var logs []model.WalletLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error
	return logs, err
}
