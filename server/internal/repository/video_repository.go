package repository

import (
	"context"
	"errors"
	"time"

	"agi-platform/server/internal/model"

	"gorm.io/gorm"
)

type VideoRepository interface {
	ListModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error)
	ListAllModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error)
	FindModelByCode(ctx context.Context, code string) (*model.VideoModel, error)
	CreateModel(ctx context.Context, videoModel *model.VideoModel) error
	UpdateModel(ctx context.Context, id uint64, values map[string]interface{}) error
	DeleteModel(ctx context.Context, tx Tx, id uint64) error
	PickRoute(ctx context.Context, modelID uint64) (*model.VideoModelRoute, error)
	ListRoutes(ctx context.Context, modelID uint64) ([]model.VideoModelRoute, error)
	CreateRoute(ctx context.Context, route *model.VideoModelRoute) error
	UpdateRoute(ctx context.Context, id uint64, values map[string]interface{}) error
	DeleteRoutesByModelID(ctx context.Context, tx Tx, modelID uint64) error
	CreateTask(ctx context.Context, tx Tx, task *model.VideoTask) error
	FindTaskByTaskNo(ctx context.Context, taskNo string) (*model.VideoTask, error)
	ListTasks(ctx context.Context, limit int, offset int) ([]model.VideoTask, error)
	ListTasksFiltered(ctx context.Context, filter TaskFilter, limit int, offset int) ([]model.VideoTask, error)
	ListTasksByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.VideoTask, error)
	MarkStaleRunningTimeout(ctx context.Context, cutoff time.Time, message string) error
	CountTasksByModelID(ctx context.Context, modelID uint64) (int64, error)
	UpdateTaskStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error
	CreateAsset(ctx context.Context, tx Tx, asset *model.VideoAsset) error
	ListAssetsByTaskID(ctx context.Context, taskID uint64) ([]model.VideoAsset, error)
}

type GormVideoRepository struct {
	db *gorm.DB
}

var videoTaskActiveStatuses = []string{model.VideoTaskStatusPending, model.VideoTaskStatusRunning}

func NewGormVideoRepository(db *gorm.DB) *GormVideoRepository {
	return &GormVideoRepository{db: db}
}

func (r *GormVideoRepository) ListModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error) {
	var models []model.VideoModel
	err := r.db.WithContext(ctx).
		Where("enabled = ?", true).
		Order("sort_order ASC, id ASC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	return models, err
}

func (r *GormVideoRepository) ListAllModels(ctx context.Context, limit int, offset int) ([]model.VideoModel, error) {
	var models []model.VideoModel
	err := r.db.WithContext(ctx).
		Order("sort_order ASC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&models).Error
	return models, err
}

func (r *GormVideoRepository) FindModelByCode(ctx context.Context, code string) (*model.VideoModel, error) {
	var videoModel model.VideoModel
	err := r.db.WithContext(ctx).
		Where("code = ? AND enabled = ?", code, true).
		First(&videoModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &videoModel, nil
}

func (r *GormVideoRepository) CreateModel(ctx context.Context, videoModel *model.VideoModel) error {
	return r.db.WithContext(ctx).Create(videoModel).Error
}

func (r *GormVideoRepository) UpdateModel(ctx context.Context, id uint64, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.VideoModel{}).
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

func (r *GormVideoRepository) DeleteModel(ctx context.Context, tx Tx, id uint64) error {
	result := r.dbOrTx(tx).WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.VideoModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormVideoRepository) PickRoute(ctx context.Context, modelID uint64) (*model.VideoModelRoute, error) {
	var route model.VideoModelRoute
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

func (r *GormVideoRepository) ListRoutes(ctx context.Context, modelID uint64) ([]model.VideoModelRoute, error) {
	var routes []model.VideoModelRoute
	err := r.db.WithContext(ctx).
		Where("model_id = ?", modelID).
		Order("priority ASC, id DESC").
		Find(&routes).Error
	return routes, err
}

func (r *GormVideoRepository) CreateRoute(ctx context.Context, route *model.VideoModelRoute) error {
	return r.db.WithContext(ctx).Create(route).Error
}

func (r *GormVideoRepository) UpdateRoute(ctx context.Context, id uint64, values map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&model.VideoModelRoute{}).
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

func (r *GormVideoRepository) DeleteRoutesByModelID(ctx context.Context, tx Tx, modelID uint64) error {
	return r.dbOrTx(tx).WithContext(ctx).
		Where("model_id = ?", modelID).
		Delete(&model.VideoModelRoute{}).Error
}

func (r *GormVideoRepository) CreateTask(ctx context.Context, tx Tx, task *model.VideoTask) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(task).Error
}

func (r *GormVideoRepository) FindTaskByTaskNo(ctx context.Context, taskNo string) (*model.VideoTask, error) {
	var task model.VideoTask
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

func (r *GormVideoRepository) ListTasks(ctx context.Context, limit int, offset int) ([]model.VideoTask, error) {
	return r.ListTasksFiltered(ctx, TaskFilter{}, limit, offset)
}

func (r *GormVideoRepository) ListTasksFiltered(ctx context.Context, filter TaskFilter, limit int, offset int) ([]model.VideoTask, error) {
	var tasks []model.VideoTask
	query := applyTaskFilter(r.db.WithContext(ctx).Model(&model.VideoTask{}), filter)
	err := query.
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *GormVideoRepository) ListTasksByUserID(ctx context.Context, userID uint64, limit int, offset int) ([]model.VideoTask, error) {
	var tasks []model.VideoTask
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error
	return tasks, err
}

func (r *GormVideoRepository) MarkStaleRunningTimeout(ctx context.Context, cutoff time.Time, message string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&model.VideoTask{}).
		Where("status IN ?", videoTaskActiveStatuses).
		Where("created_at < ?", cutoff).
		Updates(map[string]interface{}{
			"status":        model.VideoTaskStatusTimeout,
			"progress":      100,
			"error_message": message,
			"completed_at":  now,
		}).Error
}

func (r *GormVideoRepository) CountTasksByModelID(ctx context.Context, modelID uint64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.VideoTask{}).
		Where("model_id = ?", modelID).
		Count(&count).Error
	return count, err
}

func (r *GormVideoRepository) UpdateTaskStatus(ctx context.Context, tx Tx, taskID uint64, status string, progress int, values map[string]interface{}) error {
	updates := map[string]interface{}{"status": status, "progress": progress}
	for key, value := range values {
		updates[key] = value
	}
	result := r.dbOrTx(tx).WithContext(ctx).
		Model(&model.VideoTask{}).
		Where("id = ?", taskID).
		Where("status IN ?", videoTaskActiveStatuses).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *GormVideoRepository) CreateAsset(ctx context.Context, tx Tx, asset *model.VideoAsset) error {
	return r.dbOrTx(tx).WithContext(ctx).Create(asset).Error
}

func (r *GormVideoRepository) ListAssetsByTaskID(ctx context.Context, taskID uint64) ([]model.VideoAsset, error) {
	var assets []model.VideoAsset
	err := r.db.WithContext(ctx).
		Where("task_id = ? AND deleted_at IS NULL", taskID).
		Order("id ASC").
		Find(&assets).Error
	return assets, err
}

func (r *GormVideoRepository) dbOrTx(tx Tx) *gorm.DB {
	if db := dbFromTx(tx); db != nil {
		return db
	}
	return r.db
}
