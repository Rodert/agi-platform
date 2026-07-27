package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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
	modelIDs := make([]int64, 0, len(discovered))
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
		modelIDs = append(modelIDs, aiModel.ID)
	}
	if err := s.channelModelRepo.DeleteNotIn(channel.ID, modelIDs); err != nil {
		return nil, err
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
	if item.Name == "" || (item.Type != "image" && item.Type != "video" && item.Type != "text") {
		return nil, fmt.Errorf("模型名称或类型无效")
	}
	aiModel, err := s.modelRepo.FindByNameAny(item.Name)
	if err == nil {
		if aiModel.Type != item.Type {
			if !isLegacyGrokImageVideo(aiModel.Type, item, provider) {
				return nil, fmt.Errorf("模型 %s 的类型与现有目录不一致", item.Name)
			}
			// Versions before v0.1.14 classified grok-image-video as an image
			// model. Correct that one known historical record during sync.
			aiModel.Type = "video"
			aiModel.ParamsConfig = grokVideoParamsConfig(item.Name)
			if err := s.modelRepo.Update(aiModel); err != nil {
				return nil, err
			}
		}
		if provider == "grok" && item.Type == "video" && isEmptyParamsConfig(aiModel.ParamsConfig) {
			aiModel.ParamsConfig = grokVideoParamsConfig(item.Name)
			if err := s.modelRepo.Update(aiModel); err != nil {
				return nil, err
			}
		}
		if provider == "gemini" && item.Type == "image" && isEmptyParamsConfig(aiModel.ParamsConfig) {
			aiModel.ParamsConfig = geminiImageParamsConfig(item.Name)
			if err := s.modelRepo.Update(aiModel); err != nil {
				return nil, err
			}
		}
		return aiModel, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}
	paramsConfig := datatypes.JSON([]byte("{}"))
	if provider == "grok" && item.Type == "video" {
		paramsConfig = grokVideoParamsConfig(item.Name)
	}
	if provider == "gemini" && item.Type == "image" {
		paramsConfig = geminiImageParamsConfig(item.Name)
	}
	aiModel = &model.AIModel{
		Name: item.Name, DisplayName: item.Name, Type: item.Type, Provider: provider,
		Cost: 100, APIConfig: datatypes.JSON([]byte("{}")), ParamsConfig: paramsConfig,
		IsActive: true,
	}
	if err := s.modelRepo.Create(aiModel); err != nil {
		return nil, err
	}
	return aiModel, nil
}

func isLegacyGrokImageVideo(existingType string, item adapter.DiscoveredModel, provider string) bool {
	return provider == "grok" && item.Name == "grok-image-video" && existingType == "image" && item.Type == "video"
}

func geminiImageParamsConfig(name string) datatypes.JSON {
	// Gemini 3 Pro Image accepts 1K, 2K and 4K output sizes. Flash Image
	// returns 1K output, so it deliberately has no misleading quality picker.
	ratios := []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9"}
	params := map[string]interface{}{
		"ratio": map[string]interface{}{"label": "画面比例", "type": "select", "default": "1:1", "options": geminiOptions(ratios)},
	}
	if strings.Contains(strings.ToLower(name), "pro-image") {
		params["ratio"] = map[string]interface{}{"label": "画面比例", "type": "select", "default": "1:1", "options": geminiOptions(append(ratios, "21:9"))}
		params["resolution"] = map[string]interface{}{
			"label": "清晰度", "type": "select", "default": "1K",
			"options": []map[string]interface{}{
				{"value": "1K", "label": "1K"},
				{"value": "2K", "label": "2K 高清", "extra_cost": 1},
				{"value": "4K", "label": "4K 超清", "extra_cost": 2},
			},
		}
	}
	encoded, _ := json.Marshal(params)
	return datatypes.JSON(encoded)
}

func geminiOptions(values []string) []map[string]interface{} {
	options := make([]map[string]interface{}, 0, len(values))
	for _, value := range values {
		options = append(options, map[string]interface{}{"value": value, "label": value})
	}
	return options
}

func isEmptyParamsConfig(value datatypes.JSON) bool {
	text := string(value)
	return text == "" || text == "{}" || text == "null"
}

func grokVideoParamsConfig(name string) datatypes.JSON {
	ratioOptions := []string{"1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3"}
	durationOptions := []string{"6", "10", "15"}
	if name == "grok-video-1.5" || name == "grok-video-1.5-1080p" {
		ratioOptions = []string{"16:9", "9:16"}
		durationOptions = []string{"4", "6", "8", "10", "12", "15"}
	}
	if name == "grok-video-1.5fast" {
		ratioOptions = []string{"16:9", "9:16"}
		durationOptions = []string{"6", "10"}
	}
	optionList := func(values []string, suffix string) []map[string]interface{} {
		result := make([]map[string]interface{}, 0, len(values))
		for _, value := range values {
			result = append(result, map[string]interface{}{"value": value, "label": value + suffix})
		}
		return result
	}
	params, _ := json.Marshal(map[string]interface{}{
		"ratio":      map[string]interface{}{"label": "画面比例", "type": "select", "default": ratioOptions[0], "options": optionList(ratioOptions, "")},
		"resolution": map[string]interface{}{"label": "清晰度", "type": "select", "default": "720p", "options": optionList([]string{"720p", "480p"}, "")},
		"duration":   map[string]interface{}{"label": "视频时长", "type": "select", "default": durationOptions[0], "options": optionList(durationOptions, " 秒")},
	})
	return datatypes.JSON(params)
}
