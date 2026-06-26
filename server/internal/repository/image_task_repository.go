package repository

import (
	"context"
	"errors"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type ImageTaskRepository interface {
	Create(ctx context.Context, tx Tx, task *model.ImageTask) error
	FindByTaskNo(ctx context.Context, taskNo string) (*model.ImageTask, error)
	List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error)
	UpdateStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error
	CreateAsset(ctx context.Context, tx Tx, asset *model.ImageAsset) error
	ListAssetsByTaskID(ctx context.Context, taskID uint64) ([]model.ImageAsset, error)
}

type GormImageTaskRepository struct {
	db *gorm.DB
}

func NewGormImageTaskRepository(db *gorm.DB) *GormImageTaskRepository {
	return &GormImageTaskRepository{db: db}
}

func (r *GormImageTaskRepository) Create(ctx context.Context, tx Tx, task *model.ImageTask) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(task).Error
}

func (r *GormImageTaskRepository) FindByTaskNo(ctx context.Context, taskNo string) (*model.ImageTask, error) {
	var task model.ImageTask
	err := r.db.WithContext(ctx).
		Where("task_no = ?", taskNo).
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *GormImageTaskRepository) List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error) {
	var tasks []model.ImageTask
	err := r.db.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *GormImageTaskRepository) UpdateStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	for key, value := range values {
		updates[key] = value
	}

	return r.dbOrTx(tx).WithContext(ctx).
		Model(&model.ImageTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

func (r *GormImageTaskRepository) CreateAsset(ctx context.Context, tx Tx, asset *model.ImageAsset) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(asset).Error
}

func (r *GormImageTaskRepository) ListAssetsByTaskID(ctx context.Context, taskID uint64) ([]model.ImageAsset, error) {
	var assets []model.ImageAsset
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND deleted_at IS NULL", taskID).
		Order("id ASC").
		Find(&assets).Error
	return assets, err
}

func (r *GormImageTaskRepository) dbOrTx(tx Tx) *gorm.DB {
	if db := dbFromTx(tx); db != nil {
		return db
	}
	return r.db
}
