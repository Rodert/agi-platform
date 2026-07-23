package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create 创建任务
func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

// FindByID 根据ID查找任务
func (r *TaskRepository) FindByID(id int64) (*model.Task, error) {
	var task model.Task
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// FindByUserID 查找用户的任务列表
func (r *TaskRepository) FindByUserID(userID int64, status, taskType string, page, pageSize int) ([]*model.Task, int64, error) {
	var tasks []*model.Task
	var total int64

	query := r.db.Model(&model.Task{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&tasks).Error

	return tasks, total, err
}

func (r *TaskRepository) FindAdminTasks(keyword, status, taskType string, page, pageSize int) ([]*model.Task, int64, error) {
	query := r.db.Model(&model.Task{}).Joins("LEFT JOIN users ON users.id = tasks.user_id")
	if keyword != "" {
		query = query.Where("users.name LIKE ? OR users.email LIKE ? OR tasks.model_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if status != "" {
		query = query.Where("tasks.status = ?", status)
	}
	if taskType != "" {
		query = query.Where("tasks.type = ?", taskType)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tasks []*model.Task
	err := query.Preload("User").Preload("Channel").Preload("Request").Preload("Attempts", func(db *gorm.DB) *gorm.DB {
		return db.Order("attempt ASC")
	}).Preload("Assets", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Order("tasks.created_at DESC, tasks.id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

// Update 更新任务
func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// StartAttempt atomically advances the execution count and creates its audit row.
func (r *TaskRepository) StartAttempt(task *model.Task) (*model.TaskAttempt, error) {
	now := time.Now()
	attempt := &model.TaskAttempt{TaskID: task.ID, Attempt: task.AttemptCount + 1, Status: "processing", StartedAt: now}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"status": "processing", "progress": 10, "attempt_count": gorm.Expr("attempt_count + 1"), "updated_at": now,
		}).Error; err != nil {
			return err
		}
		return tx.Create(attempt).Error
	})
	if err != nil {
		return nil, err
	}
	task.AttemptCount = attempt.Attempt
	task.Status, task.Progress, task.UpdatedAt = "processing", 10, now
	return attempt, nil
}

func (r *TaskRepository) CompleteAttempt(attempt *model.TaskAttempt, status, errorMsg string) error {
	now := time.Now()
	attempt.Status, attempt.ErrorMsg, attempt.CompletedAt = status, errorMsg, &now
	return r.db.Save(attempt).Error
}

func (r *TaskRepository) MarkRetryQueued(taskID int64) error {
	now := time.Now()
	return r.db.Model(&model.Task{}).Where("id = ?", taskID).Updates(map[string]interface{}{
		"status": "queued", "progress": 0, "error_msg": "", "completed_at": nil, "last_retry_at": now, "updated_at": now,
	}).Error
}

// CountProcessingTasks 统计用户正在处理的任务数
func (r *TaskRepository) CountProcessingTasks(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Task{}).
		Where("user_id = ? AND status IN (?)", userID, []string{"queued", "processing", "uploading"}).
		Count(&count).Error
	return count, err
}

// GenerationRequestRepository 生成请求仓库
type GenerationRequestRepository struct {
	db *gorm.DB
}

func NewGenerationRequestRepository(db *gorm.DB) *GenerationRequestRepository {
	return &GenerationRequestRepository{db: db}
}

// Create 创建生成请求
func (r *GenerationRequestRepository) Create(req *model.GenerationRequest) error {
	return r.db.Create(req).Error
}
