package repository

import (
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

// Update 更新任务
func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// CountProcessingTasks 统计用户正在处理的任务数
func (r *TaskRepository) CountProcessingTasks(userID int64) (int64, error) {
	var count int64
	err := r.db.Model(&model.Task{}).
		Where("user_id = ? AND status IN (?)", userID, []string{"queued", "processing"}).
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
