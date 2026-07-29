package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
	"time"
)

type ConfigRepository struct {
	db *gorm.DB
}

func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// GetEmailConfig 获取邮箱配置（只有一条记录，ID=1）
func (r *ConfigRepository) GetEmailConfig() (*model.EmailConfig, error) {
	var config model.EmailConfig
	err := r.db.First(&config, 1).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// UpdateEmailConfig 更新邮箱配置
func (r *ConfigRepository) UpdateEmailConfig(config *model.EmailConfig) error {
	config.ID = 1 // 固定ID
	return r.db.Save(config).Error
}

// GetSystemConfig 获取系统配置
func (r *ConfigRepository) GetSystemConfig(key string) (*model.SystemConfig, error) {
	var config model.SystemConfig
	err := r.db.Where("`key` = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// GetSystemConfigsByCategory 根据分类获取配置
func (r *ConfigRepository) GetSystemConfigsByCategory(category string) ([]*model.SystemConfig, error) {
	var configs []*model.SystemConfig
	err := r.db.Where("category = ?", category).Find(&configs).Error
	return configs, err
}

// UpdateSystemConfig 更新系统配置
func (r *ConfigRepository) UpdateSystemConfig(config *model.SystemConfig) error {
	return r.db.Save(config).Error
}

func (r *ConfigRepository) UpsertSystemConfig(key, value, configType, category, description string) error {
	var config model.SystemConfig
	err := r.db.Where("`key` = ?", key).First(&config).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(&model.SystemConfig{Key: key, Value: value, Type: configType, Category: category, Description: description, UpdatedAt: time.Now()}).Error
	}
	if err != nil {
		return err
	}
	config.Value = value
	config.Type = configType
	config.Category = category
	config.Description = description
	config.UpdatedAt = time.Now()
	return r.db.Save(&config).Error
}

func (r *ConfigRepository) GetSystemConfigValue(key, fallback string) (string, error) {
	config, err := r.GetSystemConfig(key)
	if err == gorm.ErrRecordNotFound {
		return fallback, nil
	}
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

// GetTaskConfig loads the singleton task policy. The fallback keeps older
// deployments operational until the migration is applied.
func (r *ConfigRepository) GetTaskConfig() (*model.TaskConfig, error) {
	config := &model.TaskConfig{}
	err := r.db.First(config, 1).Error
	if err == gorm.ErrRecordNotFound {
		return &model.TaskConfig{ID: 1, MaxActiveTasks: 50, PromptMaxLength: 5000, MaxRetryAttempts: 0, ImageConcurrency: 8, VideoConcurrency: 2}, nil
	}
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (r *ConfigRepository) UpdateTaskConfig(config *model.TaskConfig) error {
	config.ID = 1
	config.UpdatedAt = time.Now()
	return r.db.Save(config).Error
}

func (r *ConfigRepository) GetPromptOptimizationConfig() (*model.PromptOptimizationConfig, error) {
	config := &model.PromptOptimizationConfig{}
	err := r.db.First(config, 1).Error
	if err == gorm.ErrRecordNotFound {
		return &model.PromptOptimizationConfig{ID: 1, MaxInputLength: 5000, RateLimitPerMinute: 5}, nil
	}
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (r *ConfigRepository) UpdatePromptOptimizationConfig(config *model.PromptOptimizationConfig) error {
	config.ID = 1
	config.UpdatedAt = time.Now()
	return r.db.Save(config).Error
}

// GetCategories 获取分类列表
func (r *ConfigRepository) GetCategories(categoryType string) ([]*model.Category, error) {
	var categories []*model.Category
	query := r.db.Where("is_active = ?", true)
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}
	err := query.Order("sort_order").Find(&categories).Error
	return categories, err
}
