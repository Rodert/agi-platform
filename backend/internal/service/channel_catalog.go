package service

import (
	"context"
	"fmt"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/worker/adapter"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChannelCatalogService coordinates channel discovery with the model catalog.
// It is the only application service allowed to depend on both modules.
type ChannelCatalogService struct {
	channelRepo      *repository.AIProviderAccountRepository
	modelRepo        *repository.AIModelRepository
	channelModelRepo *repository.ChannelModelRepository
}

func NewChannelCatalogService(
	channelRepo *repository.AIProviderAccountRepository,
	modelRepo *repository.AIModelRepository,
	channelModelRepo *repository.ChannelModelRepository,
) *ChannelCatalogService {
	return &ChannelCatalogService{
		channelRepo: channelRepo, modelRepo: modelRepo, channelModelRepo: channelModelRepo,
	}
}

func (s *ChannelCatalogService) Sync(ctx context.Context, channelID int64) ([]*model.ChannelModel, error) {
	channel, err := s.channelRepo.Find(channelID)
	if err != nil {
		return nil, err
	}
	discovered, err := adapter.DiscoverModels(ctx, channel)
	if err != nil {
		return nil, err
	}
	channel.HealthStatus = "healthy"
	if err := s.channelRepo.Update(channel); err != nil {
		return nil, err
	}

	bindings := make([]*model.ChannelModel, 0, len(discovered))
	for _, item := range discovered {
		aiModel, err := s.findOrCreateModel(item, channel.Provider)
		if err != nil {
			return nil, err
		}
		binding, err := s.channelModelRepo.Upsert(channel.ID, aiModel.ID, true)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, binding)
	}
	return bindings, nil
}

func (s *ChannelCatalogService) Bind(channelID int64, name, modelType string, isActive bool) (*model.ChannelModel, error) {
	channel, err := s.channelRepo.Find(channelID)
	if err != nil {
		return nil, err
	}
	aiModel, err := s.findOrCreateModel(adapter.DiscoveredModel{Name: name, Type: modelType}, channel.Provider)
	if err != nil {
		return nil, err
	}
	return s.channelModelRepo.Upsert(channel.ID, aiModel.ID, isActive)
}

func (s *ChannelCatalogService) findOrCreateModel(item adapter.DiscoveredModel, provider string) (*model.AIModel, error) {
	if item.Name == "" || (item.Type != "image" && item.Type != "video") {
		return nil, fmt.Errorf("模型名称或类型无效")
	}
	aiModel, err := s.modelRepo.FindByNameAny(item.Name)
	if err == nil {
		if aiModel.Type != item.Type {
			return nil, fmt.Errorf("模型 %s 的类型与现有目录不一致", item.Name)
		}
		return aiModel, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	aiModel = &model.AIModel{
		Name: item.Name, DisplayName: item.Name, Type: item.Type, Provider: provider,
		Cost: 0, APIConfig: datatypes.JSON([]byte("{}")), ParamsConfig: datatypes.JSON([]byte("{}")),
		IsActive: true,
	}
	if err := s.modelRepo.Create(aiModel); err != nil {
		return nil, err
	}
	return aiModel, nil
}
