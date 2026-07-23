package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type StorageConfigRepository struct {
	db *gorm.DB
}

func NewStorageConfigRepository(db *gorm.DB) *StorageConfigRepository {
	return &StorageConfigRepository{db: db}
}

// GetStorageConfigs 获取所有存储配置
func (r *StorageConfigRepository) GetStorageConfigs() ([]*model.StorageConfig, error) {
	var configs []*model.StorageConfig
	err := r.db.Order("id DESC").Find(&configs).Error
	return configs, err
}

// GetEnabledConfig 获取启用的存储配置
func (r *StorageConfigRepository) GetEnabledConfig() (*model.StorageConfig, error) {
	var config model.StorageConfig
	err := r.db.Where("is_enabled = ?", true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// Create 创建存储配置
func (r *StorageConfigRepository) Create(config *model.StorageConfig) error {
	return r.db.Create(config).Error
}

// Update 更新存储配置
func (r *StorageConfigRepository) Update(config *model.StorageConfig) error {
	return r.db.Save(config).Error
}

// EnableConfig globally selects one storage configuration in a transaction.
func (r *StorageConfigRepository) EnableConfig(id int64) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Only one storage backend may receive generated and uploaded files.
		if err := tx.Model(&model.StorageConfig{}).Where("is_enabled = ?", true).
			Update("is_enabled", false).Error; err != nil {
			return err
		}

		// 启用指定配置
		return tx.Model(&model.StorageConfig{}).Where("id = ?", id).
			Update("is_enabled", true).Error
	})
}

// Delete 删除存储配置
func (r *StorageConfigRepository) Delete(id int64) error {
	return r.db.Delete(&model.StorageConfig{}, id).Error
}

// FindByID 根据ID查找配置
func (r *StorageConfigRepository) FindByID(id int64) (*model.StorageConfig, error) {
	var config model.StorageConfig
	err := r.db.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
