package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type WorkRepository struct {
	db *gorm.DB
}

func NewWorkRepository(db *gorm.DB) *WorkRepository {
	return &WorkRepository{db: db}
}

// Create 创建作品
func (r *WorkRepository) Create(work *model.Work) error {
	return r.db.Create(work).Error
}

// FindByID 根据ID查找作品
func (r *WorkRepository) FindByID(id int64) (*model.Work, error) {
	var work model.Work
	err := r.db.First(&work, id).Error
	if err != nil {
		return nil, err
	}
	return &work, nil
}

// FindByTaskID 根据任务ID查找作品
func (r *WorkRepository) FindByTaskID(taskID int64) (*model.Work, error) {
	var work model.Work
	err := r.db.Where("task_id = ?", taskID).First(&work).Error
	if err != nil {
		return nil, err
	}
	return &work, nil
}

// FindList 查找作品列表
func (r *WorkRepository) FindList(category, workType string, userID int64, page, pageSize int) ([]*model.Work, int64, error) {
	var works []*model.Work
	var total int64

	query := r.db.Model(&model.Work{}).Where("audit_status = ?", "approved")

	if category != "" && category != "全部" {
		query = query.Where("category = ?", category)
	}
	if workType != "" {
		query = query.Where("type = ?", workType)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("published_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&works).Error

	return works, total, err
}

// FindUserList returns every lifecycle state for its owner. Unlike the public
// feed, it intentionally includes pending and rejected submissions.
func (r *WorkRepository) FindUserList(userID int64, page, pageSize int) ([]*model.Work, int64, error) {
	var works []*model.Work
	var total int64
	query := r.db.Model(&model.Work{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&works).Error
	return works, total, err
}

// Update 更新作品
func (r *WorkRepository) Update(work *model.Work) error {
	return r.db.Save(work).Error
}

// IncrementViews 增加浏览量
func (r *WorkRepository) IncrementViews(id int64) error {
	return r.db.Model(&model.Work{}).Where("id = ?", id).
		UpdateColumn("views_count", gorm.Expr("views_count + ?", 1)).Error
}

// IsLiked 是否已点赞
func (r *WorkRepository) IsLiked(userID, workID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.WorkLike{}).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Count(&count).Error
	return count > 0, err
}

// IsCollected 是否已收藏
func (r *WorkRepository) IsCollected(userID, workID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.WorkCollect{}).
		Where("user_id = ? AND work_id = ?", userID, workID).
		Count(&count).Error
	return count > 0, err
}

// Like 点赞
func (r *WorkRepository) Like(userID, workID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建点赞记录
		like := &model.WorkLike{
			UserID: userID,
			WorkID: workID,
		}
		if err := tx.Create(like).Error; err != nil {
			return err
		}

		// 增加点赞数
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("likes_count", gorm.Expr("likes_count + ?", 1)).Error
	})
}

// Unlike 取消点赞
func (r *WorkRepository) Unlike(userID, workID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除点赞记录
		if err := tx.Where("user_id = ? AND work_id = ?", userID, workID).
			Delete(&model.WorkLike{}).Error; err != nil {
			return err
		}

		// 减少点赞数
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("likes_count", gorm.Expr("likes_count - ?", 1)).Error
	})
}

// Collect 收藏
func (r *WorkRepository) Collect(userID, workID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 创建收藏记录
		collect := &model.WorkCollect{
			UserID: userID,
			WorkID: workID,
		}
		if err := tx.Create(collect).Error; err != nil {
			return err
		}

		// 增加收藏数
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("collects_count", gorm.Expr("collects_count + ?", 1)).Error
	})
}

// Uncollect 取消收藏
func (r *WorkRepository) Uncollect(userID, workID int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除收藏记录
		if err := tx.Where("user_id = ? AND work_id = ?", userID, workID).
			Delete(&model.WorkCollect{}).Error; err != nil {
			return err
		}

		// 减少收藏数
		return tx.Model(&model.Work{}).Where("id = ?", workID).
			UpdateColumn("collects_count", gorm.Expr("collects_count - ?", 1)).Error
	})
}

// CreateAudit 创建审核记录
func (r *WorkRepository) CreateAudit(audit *model.WorkAudit) error {
	return r.db.Create(audit).Error
}

// FindPendingWorks 查找待审核作品
func (r *WorkRepository) FindPendingWorks(page, pageSize int) ([]*model.Work, int64, error) {
	var works []*model.Work
	var total int64

	query := r.db.Model(&model.Work{}).Where("audit_status = ?", "pending")

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&works).Error

	return works, total, err
}

// FindAdminList returns every work lifecycle state for administrative management.
func (r *WorkRepository) FindAdminList(status string, page, pageSize int) ([]*model.Work, int64, error) {
	var works []*model.Work
	var total int64
	query := r.db.Model(&model.Work{})
	if status != "" {
		query = query.Where("audit_status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("updated_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&works).Error
	return works, total, err
}
