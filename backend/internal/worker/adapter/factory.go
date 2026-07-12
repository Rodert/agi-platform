package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/javapub/agi-platform-backend/internal/model"
)

// GetAdapter 根据 AI 模型获取适配器
func GetAdapter(aiModel *model.AIModel) (Adapter, error) {
	// 解析 API 配置
	var apiConfig map[string]interface{}
	if err := json.Unmarshal(aiModel.APIConfig, &apiConfig); err != nil {
		return nil, fmt.Errorf("解析 API 配置失败: %w", err)
	}
	if account := aiModel.ProviderAccount; account != nil {
		apiConfig["api_url"] = account.APIURL
		apiConfig["api_key"] = account.APIKey
		if len(account.ExtraConfig) > 0 {
			var extra map[string]interface{}
			if json.Unmarshal(account.ExtraConfig, &extra) == nil {
				for key, value := range extra { apiConfig[key] = value }
			}
		}
	}

	// 根据提供商返回对应的适配器
	switch aiModel.Provider {
	case "openai":
		return NewOpenAIAdapter(apiConfig), nil
	case "jimeng":
		return NewJimengAdapter(apiConfig), nil
	case "demo":
		return NewDemoAdapter(), nil
	default:
		return nil, fmt.Errorf("不支持的提供商: %s", aiModel.Provider)
	}
}

// DemoAdapter 演示适配器（用于测试）
type DemoAdapter struct{}

func NewDemoAdapter() *DemoAdapter {
	return &DemoAdapter{}
}

func (a *DemoAdapter) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// 模拟生成延迟
	// time.Sleep(3 * time.Second)

	// 返回测试图片
	return &GenerateResponse{
		ImageURL:     "https://picsum.photos/1024/1024",
		ThumbnailURL: "https://picsum.photos/512/512",
	}, nil
}

// OpenAIAdapter OpenAI 适配器
type OpenAIAdapter struct {
	apiConfig map[string]interface{}
}

func NewOpenAIAdapter(apiConfig map[string]interface{}) *OpenAIAdapter {
	return &OpenAIAdapter{apiConfig: apiConfig}
}

func (a *OpenAIAdapter) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// TODO: 调用 OpenAI DALL-E API
	// apiURL := a.apiConfig["api_url"].(string)
	// apiKey := a.apiConfig["api_key"].(string)

	// 临时返回演示数据
	return &GenerateResponse{
		ImageURL: "https://picsum.photos/1024/1024",
	}, nil
}

// JimengAdapter 即梦适配器
type JimengAdapter struct {
	apiConfig map[string]interface{}
}

func NewJimengAdapter(apiConfig map[string]interface{}) *JimengAdapter {
	return &JimengAdapter{apiConfig: apiConfig}
}

func (a *JimengAdapter) Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// TODO: 调用即梦 API
	// apiURL := a.apiConfig["api_url"].(string)
	// apiKey := a.apiConfig["api_key"].(string)

	// 临时返回演示数据
	return &GenerateResponse{
		ImageURL: "https://picsum.photos/1024/1024",
	}, nil
}
