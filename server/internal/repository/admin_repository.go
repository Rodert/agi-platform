package repository

import (
	"context"
	"errors"
	"time"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type AdminRepository interface {
	FindByID(ctx context.Context, id uint64) (*model.AdminUser, error)
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	TouchLastLogin(ctx context.Context, id uint64) error
}

type GormAdminRepository struct {
	db *gorm.DB
}

func NewGormAdminRepository(db *gorm.DB) *GormAdminRepository {
	return &GormAdminRepository{db: db}
}

func (r *GormAdminRepository) FindByID(ctx context.Context, id uint64) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).First(&admin, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminRepository) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	if err := r.db.WithContext(ctx).Where("username = ?", username).First(&admin).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &admin, nil
}

func (r *GormAdminRepository) TouchLastLogin(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("id = ?", id).
		Update("last_login_at", time.Now()).
		Error
}
