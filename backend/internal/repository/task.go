package repository

import (
	"errors"
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

// FindPendingProviderTasks returns upstream jobs that were submitted before a
// worker stopped. Re-enqueuing them resumes polling without submitting again.
func (r *TaskRepository) FindPendingProviderTasks() ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.Preload("Request").
		Where("provider_task_id <> '' AND status IN (?)", []string{"processing", "polling"}).
		Order("updated_at ASC, id ASC").
		Find(&tasks).Error
	return tasks, err
}

// FindQueuedTasks returns tasks that were persisted but not picked up before a
// worker restart. Re-publishing them is safe because no upstream submission began.
func (r *TaskRepository) FindQueuedTasks() ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.Preload("Request").Where("status = ?", "queued").Order("created_at ASC, id ASC").Find(&tasks).Error
	return tasks, err
}

// FindInterruptedSubmissions finds the only unsafe restart window: the request
// may have reached an upstream that cannot deduplicate it, but no task ID was saved.
func (r *TaskRepository) FindInterruptedSubmissions() ([]*model.Task, error) {
	var tasks []*model.Task
	err := r.db.Preload("Attempts", func(db *gorm.DB) *gorm.DB {
		return db.Order("attempt DESC")
	}).Where("status = ? AND submission_state IN ? AND provider_task_id = ?", "processing", []string{"submitting", "submitted"}, "").Find(&tasks).Error
	return tasks, err
}

func (r *TaskRepository) FindOpenAttempt(taskID int64) (*model.TaskAttempt, error) {
	var attempt model.TaskAttempt
	err := r.db.Where("task_id = ? AND status = ?", taskID, "processing").Order("attempt DESC").First(&attempt).Error
	if err != nil {
		return nil, err
	}
	return &attempt, nil
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
		result := tx.Model(&model.Task{}).Where("id = ? AND status = ?", task.ID, "queued").Updates(map[string]interface{}{
			"status": "processing", "progress": 10, "attempt_count": gorm.Expr("attempt_count + 1"), "updated_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("任务已被其他消费者抢占")
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

// MarkSubmitting is persisted immediately before an upstream generation call.
// A restart can then distinguish an unsubmitted queued task from an uncertain call.
func (r *TaskRepository) MarkSubmitting(task *model.Task) error {
	now := time.Now()
	if err := r.db.Model(&model.Task{}).Where("id = ? AND status = ?", task.ID, "processing").Updates(map[string]interface{}{
		"submission_state": "submitting", "updated_at": now,
	}).Error; err != nil {
		return err
	}
	task.SubmissionState, task.UpdatedAt = "submitting", now
	return nil
}

func (r *TaskRepository) MarkSubmissionAccepted(task *model.Task) error {
	now := time.Now()
	if err := r.db.Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"submission_state": "submitted", "updated_at": now,
	}).Error; err != nil {
		return err
	}
	task.SubmissionState, task.UpdatedAt = "submitted", now
	return nil
}

// FailInterruptedSubmission marks an uncertain submission as failed. It is
// intentionally conditional, so a live worker can never be overwritten.
func (r *TaskRepository) FailInterruptedSubmission(task *model.Task, errorMsg string) (bool, error) {
	now := time.Now()
	updated := false
	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.Task{}).Where("id = ? AND status = ? AND submission_state IN ? AND provider_task_id = ?", task.ID, "processing", []string{"submitting", "submitted"}, "").Updates(map[string]interface{}{
			"status": "failed", "error_msg": errorMsg, "submission_state": "unknown", "updated_at": now, "completed_at": now,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		updated = true
		return tx.Model(&model.TaskAttempt{}).Where("task_id = ? AND status = ?", task.ID, "processing").Updates(map[string]interface{}{
			"status": "failed", "error_msg": errorMsg, "completed_at": now,
		}).Error
	})
	if err != nil || !updated {
		return updated, err
	}
	task.Status, task.ErrorMsg, task.SubmissionState, task.UpdatedAt, task.CompletedAt = "failed", errorMsg, "unknown", now, &now
	return true, nil
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
