package repository

import (
	"github.com/javapub/agi-platform-backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ChannelModelRepository owns channel availability. It deliberately does not
// read or alter model capability schemas.
type ChannelModelRepository struct {
	db *gorm.DB
}

func NewChannelModelRepository(db *gorm.DB) *ChannelModelRepository {
	return &ChannelModelRepository{db: db}
}

func (r *ChannelModelRepository) SelectActiveChannel(modelID int64) (*model.ChannelModel, error) {
	var binding model.ChannelModel
	err := r.db.
		Preload("Channel").
		Where("channel_models.model_id = ? AND channel_models.is_active = ?", modelID, true).
		Joins("JOIN ai_provider_accounts ON ai_provider_accounts.id = channel_models.channel_id").
		Where("ai_provider_accounts.is_active = ?", true).
		Order("ai_provider_accounts.priority ASC, channel_models.id ASC").
		First(&binding).Error
	return &binding, err
}

func (r *ChannelModelRepository) FindActiveChannel(id int64) (*model.AIProviderAccount, error) {
	var channel model.AIProviderAccount
	err := r.db.Where("id = ? AND is_active = ?", id, true).First(&channel).Error
	return &channel, err
}

func (r *ChannelModelRepository) Upsert(channelID, modelID int64, isActive bool) (*model.ChannelModel, error) {
	binding := &model.ChannelModel{ChannelID: channelID, ModelID: modelID, IsActive: isActive}
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "channel_id"}, {Name: "model_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"is_active", "updated_at"}),
	}).Create(binding).Error
	if err != nil {
		return nil, err
	}
	var result model.ChannelModel
	err = r.db.Preload("Model").First(&result, "channel_id = ? AND model_id = ?", channelID, modelID).Error
	return &result, err
}

func (r *ChannelModelRepository) UpdateStatus(channelID, modelID int64, isActive bool) error {
	return r.db.Model(&model.ChannelModel{}).
		Where("channel_id = ? AND model_id = ?", channelID, modelID).
		Update("is_active", isActive).Error
}
