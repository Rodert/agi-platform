package repository

import (
	"context"
	"errors"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type ProviderRepository interface {
	List(ctx context.Context, limit int, offset int) ([]model.Provider, error)
	ListEnabled(ctx context.Context) ([]model.Provider, error)
	FindByID(ctx context.Context, id uint64) (*model.Provider, error)
	Create(ctx context.Context, provider *model.Provider) error
	Update(ctx context.Context, id uint64, values map[string]interface{}) error
	FindKeyByID(ctx context.Context, id uint64) (*model.ProviderKey, error)
	PickActiveKey(ctx context.Context, providerID uint64) (*model.ProviderKey, error)
	ListKeys(ctx context.Context, providerID uint64) ([]model.ProviderKey, error)
	CreateKey(ctx context.Context, key *model.ProviderKey) error
	DeleteKey(ctx context.Context, id uint64) error
}

type GormProviderRepository struct {
	db *gorm.DB
}

func NewGormProviderRepository(db *gorm.DB) *GormProviderRepository {
	return &GormProviderRepository{db: db}
}

func (r *GormProviderRepository) List(ctx context.Context, limit int, offset int) ([]model.Provider, error) {
	var providers []model.Provider
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&providers).Error
	return providers, err
}

func (r *GormProviderRepository) ListEnabled(ctx context.Context) ([]model.Provider, error) {
	var providers []model.Provider
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("priority ASC, id ASC").
		Find(&providers).Error
	return providers, err
}

func (r *GormProviderRepository) FindByID(ctx context.Context, id uint64) (*model.Provider, error) {
	var provider model.Provider
	if err := r.db.WithContext(ctx).First(&provider, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &provider, nil
}

func (r *GormProviderRepository) Create(ctx context.Context, provider *model.Provider) error {
	return r.db.WithContext(ctx).Create(provider).Error
}

func (r *GormProviderRepository) Update(ctx context.Context, id uint64, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.Provider{}).
		Where("id = ?", id).
		Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormProviderRepository) FindKeyByID(ctx context.Context, id uint64) (*model.ProviderKey, error) {
	var key model.ProviderKey
	err := r.db.WithContext(ctx).
		Where("id = ? AND status = ?", id, "active").
		First(&key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *GormProviderRepository) PickActiveKey(ctx context.Context, providerID uint64) (*model.ProviderKey, error) {
	var key model.ProviderKey
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND status = ?", providerID, "active").
		Order("weight DESC, last_used_at ASC, id ASC").
		First(&key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &key, nil
}

func (r *GormProviderRepository) ListKeys(ctx context.Context, providerID uint64) ([]model.ProviderKey, error) {
	var keys []model.ProviderKey
	err := r.db.WithContext(ctx).
		Where("provider_id = ? AND deleted_at IS NULL", providerID).
		Order("id DESC").
		Find(&keys).Error
	return keys, err
}

func (r *GormProviderRepository) CreateKey(ctx context.Context, key *model.ProviderKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *GormProviderRepository) DeleteKey(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.ProviderKey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
