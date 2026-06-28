package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type ImageTaskRepository interface {
	Create(ctx context.Context, tx Tx, task *model.ImageTask) error
	FindByTaskNo(ctx context.Context, taskNo string) (*model.ImageTask, error)
	List(ctx context.Context, limit int, offset int) ([]model.ImageTask, error)
	ListFiltered(ctx context.Context, filter TaskFilter, limit int, offset int) ([]model.ImageTask, error)
	ListByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.ImageTask, error)
	MarkStaleRunningTimeout(ctx context.Context, cutoff time.Time, message string) error
	CountByModelID(ctx context.Context, modelID uint64) (int64, error)
	UpdateStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error
	CreateAsset(ctx context.Context, tx Tx, asset *model.ImageAsset) error
	ListAssetsByTaskID(ctx context.Context, taskID uint64) ([]model.ImageAsset, error)
}

type TaskFilter struct {
	ID      uint64
	TaskNo  string
	Keyword string
	Status  string
}

var imageTaskActiveStatuses = []string{model.ImageTaskStatusPending, model.ImageTaskStatusRunning}

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
	return r.ListFiltered(ctx, TaskFilter{}, limit, offset)
}

func (r *GormImageTaskRepository) ListFiltered(ctx context.Context, filter TaskFilter, limit int, offset int) ([]model.ImageTask, error) {
	var tasks []model.ImageTask
	query := applyTaskFilter(r.db.WithContext(ctx).Model(&model.ImageTask{}), filter)
	err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func applyTaskFilter(query *gorm.DB, filter TaskFilter) *gorm.DB {
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if taskNo := strings.TrimSpace(filter.TaskNo); taskNo != "" {
		query = query.Where("task_no LIKE ?", "%"+taskNo+"%")
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		query = query.Where("task_no LIKE ? OR prompt LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	return query
}

func (r *GormImageTaskRepository) ListByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.ImageTask, error) {
	var tasks []model.ImageTask
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *GormImageTaskRepository) MarkStaleRunningTimeout(ctx context.Context, cutoff time.Time, message string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.ImageTask{}).
		Where("status IN ?", imageTaskActiveStatuses).
		Where("created_at < ?", cutoff).
		Updates(map[string]interface{}{
			"status":        model.ImageTaskStatusTimeout,
			"progress":      100,
			"error_message": message,
			"completed_at":  now,
		}).Error
}

func (r *GormImageTaskRepository) CountByModelID(ctx context.Context, modelID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.ImageTask{}).
		Where("model_id = ?", modelID).
		Count(&count).Error
	return count, err
}

func (r *GormImageTaskRepository) UpdateStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}
	for key, value := range values {
		updates[key] = value
	}

	result := r.dbOrTx(tx).WithContext(ctx).
		Model(&model.ImageTask{}).
		Where("id = ?", taskID).
		Where("status IN ?", imageTaskActiveStatuses).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
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
