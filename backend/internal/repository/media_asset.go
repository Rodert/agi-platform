package repository

import (
	"time"

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

func (r *MediaAssetRepository) FindExpired(before time.Time, limit int) ([]*model.MediaAsset, error) {
	var assets []*model.MediaAsset
	err := r.db.Where("expires_at IS NOT NULL AND expires_at <= ?", before).
		Order("expires_at ASC, id ASC").Limit(limit).Find(&assets).Error
	return assets, err
}

func (r *MediaAssetRepository) Delete(id int64) error {
	return r.db.Delete(&model.MediaAsset{}, id).Error
}
