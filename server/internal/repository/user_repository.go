package repository

import (
	"context"
	"errors"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("record not found")

type UserRepository interface {
	Create(ctx context.Context, tx Tx, user *model.User) error
	FindByID(ctx context.Context, id uint64) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	List(ctx context.Context, limit int, offset int) ([]model.User, error)
	Update(ctx context.Context, tx Tx, user *model.User) error
	DeductCredits(ctx context.Context, tx Tx, userID uint64, amount int64) (*model.User, error)
	AddCredits(ctx context.Context, tx Tx, userID uint64, amount int64) (*model.User, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) Create(ctx context.Context, tx Tx, user *model.User) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uint64) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) List(ctx context.Context, limit int, offset int) ([]model.User, error) {
	var users []model.User
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, err
}

func (r *GormUserRepository) Update(ctx context.Context, tx Tx, user *model.User) error {
	result := r.dbOrTx(tx).WithContext(ctx).Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{
			"email":         user.Email,
			"phone":         user.Phone,
			"password_hash": user.PasswordHash,
			"nickname":      user.Nickname,
			"avatar_url":    user.AvatarURL,
			"status":        user.Status,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormUserRepository) DeductCredits(ctx context.Context, tx Tx, userID uint64, amount int64) (*model.User, error) {
	db := r.dbOrTx(tx).WithContext(ctx)

	result := db.Model(&model.User{}).
		Where("id = ? AND credits >= ?", userID, amount).
		UpdateColumn("credits", gorm.Expr("credits - ?", amount))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.findByID(ctx, db, userID)
}

func (r *GormUserRepository) AddCredits(ctx context.Context, tx Tx, userID uint64, amount int64) (*model.User, error) {
	db := r.dbOrTx(tx).WithContext(ctx)

	result := db.Model(&model.User{}).
		Where("id = ?", userID).
		UpdateColumn("credits", gorm.Expr("credits + ?", amount))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrNotFound
	}

	return r.findByID(ctx, db, userID)
}

func (r *GormUserRepository) findByID(ctx context.Context, db *gorm.DB, id uint64) (*model.User, error) {
	var user model.User
	if err := db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) dbOrTx(tx Tx) *gorm.DB {
	if db := dbFromTx(tx); db != nil {
		return db
	}
	return r.db
}
