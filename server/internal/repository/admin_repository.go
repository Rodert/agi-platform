package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type AdminRepository interface {
	Create(ctx context.Context, admin *model.AdminUser) error
	FindByID(ctx context.Context, id uint64) (*model.AdminUser, error)
	FindByUsername(ctx context.Context, username string) (*model.AdminUser, error)
	Update(ctx context.Context, admin *model.AdminUser) error
	UpdatePassword(ctx context.Context, id uint64, passwordHash string) error
	TouchLastLogin(ctx context.Context, id uint64) error
}

type GormAdminRepository struct {
	db *gorm.DB
}

func NewGormAdminRepository(db *gorm.DB) *GormAdminRepository {
	return &GormAdminRepository{db: db}
}

func (r *GormAdminRepository) Create(ctx context.Context, admin *model.AdminUser) error {
	return r.db.WithContext(ctx).Create(admin).Error
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

func (r *GormAdminRepository) Update(ctx context.Context, admin *model.AdminUser) error {
	values := map[string]any{
		"nickname": admin.Nickname,
		"role":     admin.Role,
		"status":   admin.Status,
	}
	if strings.TrimSpace(admin.PasswordHash) != "" {
		values["password_hash"] = admin.PasswordHash
	}
	result := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("id = ?", admin.ID).
		Updates(values)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormAdminRepository) UpdatePassword(ctx context.Context, id uint64, passwordHash string) error {
	result := r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("id = ?", id).
		Update("password_hash", passwordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormAdminRepository) TouchLastLogin(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).
		Model(&model.AdminUser{}).
		Where("id = ?", id).
		Update("last_login_at", time.Now()).
		Error
}
