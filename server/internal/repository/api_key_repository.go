package repository

import (
	"context"
	"errors"
	"time"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type APIKeyRepository interface {
	Create(ctx context.Context, tx Tx, key *model.APIKey) error
	ListByUserID(ctx context.Context, userID uint64) ([]model.APIKey, error)
	FindActiveByHash(ctx context.Context, hash string) (*model.APIKey, error)
	Revoke(ctx context.Context, userID uint64, id uint64) error
	TouchLastUsed(ctx context.Context, id uint64) error
}

type GormAPIKeyRepository struct {
	db *gorm.DB
}

func NewGormAPIKeyRepository(db *gorm.DB) *GormAPIKeyRepository {
	return &GormAPIKeyRepository{db: db}
}

func (r *GormAPIKeyRepository) Create(ctx context.Context, tx Tx, key *model.APIKey) error {
	db := r.db
	if txDB := dbFromTx(tx); txDB != nil {
		db = txDB
	}
	return db.WithContext(ctx).Create(key).Error
}

func (r *GormAPIKeyRepository) ListByUserID(ctx context.Context, userID uint64) ([]model.APIKey, error) {
	var keys []model.APIKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("id DESC").
		Find(&keys).Error
	return keys, err
}

func (r *GormAPIKeyRepository) FindActiveByHash(ctx context.Context, hash string) (*model.APIKey, error) {
	var key model.APIKey
	err := r.db.WithContext(ctx).
		Where("key_hash = ? AND status = ? AND deleted_at IS NULL", hash, "active").
		First(&key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *GormAPIKeyRepository) Revoke(ctx context.Context, userID uint64, id uint64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"status":     "revoked",
			"deleted_at": now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormAPIKeyRepository) TouchLastUsed(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.APIKey{}).
		Where("id = ?", id).
		Update("last_used_at", time.Now()).
		Error
}
