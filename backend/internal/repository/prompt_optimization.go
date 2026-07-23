package repository

import (
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type PromptOptimizationRepository struct{ db *gorm.DB }

func NewPromptOptimizationRepository(db *gorm.DB) *PromptOptimizationRepository {
	return &PromptOptimizationRepository{db: db}
}

func (r *PromptOptimizationRepository) Create(log *model.PromptOptimizationLog) error {
	return r.db.Create(log).Error
}

func (r *PromptOptimizationRepository) Update(log *model.PromptOptimizationLog) error {
	return r.db.Save(log).Error
}

func (r *PromptOptimizationRepository) UpdateTx(tx *gorm.DB, log *model.PromptOptimizationLog) error {
	return tx.Save(log).Error
}

func (r *PromptOptimizationRepository) CountUserSince(userID int64, since time.Time) (int64, error) {
	var count int64
	err := r.db.Model(&model.PromptOptimizationLog{}).Where("user_id = ? AND created_at >= ?", userID, since).Count(&count).Error
	return count, err
}

func (r *PromptOptimizationRepository) ListAdmin(page, pageSize int) ([]*model.PromptOptimizationLog, int64, error) {
	var logs []*model.PromptOptimizationLog
	var total int64
	query := r.db.Model(&model.PromptOptimizationLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("User").Preload("Channel").Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
