package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
)

type AIModelRepository struct {
	db *gorm.DB
}

func NewAIModelRepository(db *gorm.DB) *AIModelRepository {
	return &AIModelRepository{db: db}
}

// FindByName 根据名称查找模型
func (r *AIModelRepository) FindByName(name string) (*model.AIModel, error) {
	var aiModel model.AIModel
	err := r.db.
		Where("ai_models.name = ?", name).
		Where("EXISTS (SELECT 1 FROM channel_models cm JOIN ai_provider_accounts c ON c.id = cm.channel_id WHERE cm.model_id = ai_models.id AND cm.is_active = ? AND c.is_active = ?)", true, true).
		First(&aiModel).Error
	if err != nil {
		return nil, err
	}
	return &aiModel, nil
}

// GetActiveModels 获取启用的模型列表
func (r *AIModelRepository) GetActiveModels(modelType string) ([]*model.AIModel, error) {
	var models []*model.AIModel
	query := r.db.Where("EXISTS (SELECT 1 FROM channel_models cm JOIN ai_provider_accounts c ON c.id = cm.channel_id WHERE cm.model_id = ai_models.id AND cm.is_active = ? AND c.is_active = ?)", true, true)

	if modelType != "" {
		query = query.Where("ai_models.type = ?", modelType)
	}

	err := query.Order("sort_order ASC, created_at DESC").Find(&models).Error
	return models, err
}

func (r *AIModelRepository) GetAllModels() ([]*model.AIModel, error) {
	var models []*model.AIModel
	err := r.db.Order("type ASC, sort_order ASC, id ASC").Find(&models).Error
	return models, err
}

// FindByID 根据ID查找模型
func (r *AIModelRepository) FindByID(id int64) (*model.AIModel, error) {
	var aiModel model.AIModel
	err := r.db.First(&aiModel, id).Error
	if err != nil {
		return nil, err
	}
	return &aiModel, nil
}

func (r *AIModelRepository) FindByNameAny(name string) (*model.AIModel, error) {
	var aiModel model.AIModel
	err := r.db.Where("name = ?", name).First(&aiModel).Error
	return &aiModel, err
}

func (r *AIModelRepository) Create(aiModel *model.AIModel) error {
	return r.db.Create(aiModel).Error
}

// Update 更新模型
func (r *AIModelRepository) Update(model *model.AIModel) error {
	return r.db.Save(model).Error
}

// UpdateStatus 更新模型状态
func (r *AIModelRepository) UpdateStatus(id int64, isActive bool) error {
	return r.db.Model(&model.AIModel{}).Where("id = ?", id).Update("is_active", isActive).Error
}
