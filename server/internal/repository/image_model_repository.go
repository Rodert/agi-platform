package repository

import (
	"context"
	"errors"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type ImageModelRepository interface {
	List(ctx context.Context, limit int, offset int) ([]model.ImageModel, error)
	ListEnabled(ctx context.Context) ([]model.ImageModel, error)
	FindByID(ctx context.Context, id uint64) (*model.ImageModel, error)
	FindByCode(ctx context.Context, code string) (*model.ImageModel, error)
	Create(ctx context.Context, imageModel *model.ImageModel) error
	Update(ctx context.Context, id uint64, values map[string]interface{}) error
	Delete(ctx context.Context, tx Tx, id uint64) error
	PickRoute(ctx context.Context, modelID uint64) (*model.ImageModelRoute, error)
	ListRoutes(ctx context.Context, modelID uint64) ([]model.ImageModelRoute, error)
	CreateRoute(ctx context.Context, route *model.ImageModelRoute) error
	UpdateRoute(ctx context.Context, id uint64, values map[string]interface{}) error
	DeleteRoutesByModelID(ctx context.Context, tx Tx, modelID uint64) error
}

type GormImageModelRepository struct {
	db *gorm.DB
}

func NewGormImageModelRepository(db *gorm.DB) *GormImageModelRepository {
	return &GormImageModelRepository{db: db}
}

func (r *GormImageModelRepository) List(ctx context.Context, limit int, offset int) ([]model.ImageModel, error) {
	var models []model.ImageModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	return models, err
}

func (r *GormImageModelRepository) ListEnabled(ctx context.Context) ([]model.ImageModel, error) {
	var models []model.ImageModel
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").
		Find(&models).Error
	return models, err
}

func (r *GormImageModelRepository) FindByID(ctx context.Context, id uint64) (*model.ImageModel, error) {
	var imageModel model.ImageModel
	if err := r.db.WithContext(ctx).First(&imageModel, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &imageModel, nil
}

func (r *GormImageModelRepository) FindByCode(ctx context.Context, code string) (*model.ImageModel, error) {
	var imageModel model.ImageModel
	err := r.db.WithContext(ctx).
		Where("code = ? AND enabled = ?", code, true).
		First(&imageModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &imageModel, nil
}

func (r *GormImageModelRepository) Create(ctx context.Context, imageModel *model.ImageModel) error {
	return r.db.WithContext(ctx).Create(imageModel).Error
}

func (r *GormImageModelRepository) Update(ctx context.Context, id uint64, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.ImageModel{}).
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

func (r *GormImageModelRepository) Delete(ctx context.Context, tx Tx, id uint64) error {
	result := r.dbOrTx(tx).WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.ImageModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormImageModelRepository) PickRoute(ctx context.Context, modelID uint64) (*model.ImageModelRoute, error) {
	var route model.ImageModelRoute
	err := r.db.WithContext(ctx).
		Where("model_id = ? AND enabled = ?", modelID, true).
		Order("priority ASC, weight DESC, id ASC").
		First(&route).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &route, nil
}

func (r *GormImageModelRepository) ListRoutes(ctx context.Context, modelID uint64) ([]model.ImageModelRoute, error) {
	var routes []model.ImageModelRoute
	err := r.db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Order("priority ASC, id DESC").
		Find(&routes).Error
	return routes, err
}

func (r *GormImageModelRepository) CreateRoute(ctx context.Context, route *model.ImageModelRoute) error {
	return r.db.WithContext(ctx).Create(route).Error
}

func (r *GormImageModelRepository) UpdateRoute(ctx context.Context, id uint64, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.ImageModelRoute{}).
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

func (r *GormImageModelRepository) DeleteRoutesByModelID(ctx context.Context, tx Tx, modelID uint64) error {
	return r.dbOrTx(tx).WithContext(ctx).
		Where("model_id = ?", modelID).
		Delete(&model.ImageModelRoute{}).Error
}

func (r *GormImageModelRepository) dbOrTx(tx Tx) *gorm.DB {
	if db := dbFromTx(tx); db != nil {
		return db
	}
	return r.db
}
