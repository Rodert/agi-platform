package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type MediaAssetRepository struct{ db *gorm.DB }

func NewMediaAssetRepository(db *gorm.DB) *MediaAssetRepository { return &MediaAssetRepository{db: db} }

func (r *MediaAssetRepository) Create(asset *model.MediaAsset) error { return r.db.Create(asset).Error }

func (r *MediaAssetRepository) CreateTx(tx *gorm.DB, asset *model.MediaAsset) error {
	return tx.Create(asset).Error
}

func (r *MediaAssetRepository) FindByTaskAndURL(taskID int64, resourceType, publicURL string) (*model.MediaAsset, error) {
	var asset model.MediaAsset
	err := r.db.Where("task_id = ? AND resource_type = ? AND public_url = ?", taskID, resourceType, publicURL).
		Order("id DESC").First(&asset).Error
	return &asset, err
}
