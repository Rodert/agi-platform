package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
)

var factories = map[string]Factory{}

func Register(provider string, factory Factory) { factories[provider] = factory }

func init() {
	Register("openai", func(config map[string]interface{}) (Adapter, error) { return NewOpenAIAdapter(config), nil })
	Register("chatgpt", func(config map[string]interface{}) (Adapter, error) { return NewOpenAIAdapter(config), nil })
	Register("gemini", func(config map[string]interface{}) (Adapter, error) { return NewGeminiAdapter(config), nil })
	Register("jimeng", func(config map[string]interface{}) (Adapter, error) { return NewJimengAdapter(config), nil })
	Register("jimeng_international", func(config map[string]interface{}) (Adapter, error) { return NewJimengInternationalAdapter(config), nil })
	Register("grok", func(config map[string]interface{}) (Adapter, error) { return NewGrokAdapter(config), nil })
	Register("demo", func(config map[string]interface{}) (Adapter, error) { return NewDemoAdapter(), nil })
}

// GetAdapter resolves a provider implementation from the registry. Provider
// modules can be added without changing task orchestration code.
func GetAdapter(aiModel *model.AIModel, channel *model.AIProviderAccount) (Adapter, error) {
	// 解析 API 配置
	var apiConfig map[string]interface{}
	if err := json.Unmarshal(aiModel.APIConfig, &apiConfig); err != nil {
		return nil, fmt.Errorf("解析 API 配置失败: %w", err)
	}
	if channel != nil {
		apiConfig["api_url"] = channel.APIURL
		apiConfig["api_key"] = channel.APIKey
		if len(channel.ExtraConfig) > 0 {
			var extra map[string]interface{}
			if json.Unmarshal(channel.ExtraConfig, &extra) == nil {
				for key, value := range extra {
					apiConfig[key] = value
				}
			}
		}
	}

	// 根据提供商返回对应的适配器
	provider := aiModel.Provider
	if channel != nil {
		provider = channel.Provider
	}
	factory, ok := factories[provider]
	if !ok {
		return nil, fmt.Errorf("不支持的提供商: %s", provider)
	}
	return factory(apiConfig)
}

// DiscoverModels asks an upstream channel which image and video models it
// exposes. The adapter normalizes only the model name and type; user-facing
// capability schemas remain in the model catalog.
func DiscoverModels(ctx context.Context, channel *model.AIProviderAccount) ([]DiscoveredModel, error) {
	switch channel.Provider {
	case "openai", "chatgpt":
		return discoverOpenAIModels(ctx, channel)
	case "gemini":
		return discoverGeminiModels(ctx, channel)
	case "grok":
		return discoverGrokModels(ctx, channel)
	case "jimeng_international":
		return discoverJimengInternationalModels(ctx, channel)
	case "demo":
		return []DiscoveredModel{{Name: "demo-image", Type: "image"}}, nil
	default:
		return nil, fmt.Errorf("渠道 %s 尚未实现模型发现", channel.Provider)
	}
}

func discoverOpenAIModels(ctx context.Context, channel *model.AIProviderAccount) ([]DiscoveredModel, error) {
	endpoint := modelListEndpoint(channel.APIURL, "/v1/models")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+channel.APIKey)
	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := doJSON(req, &payload); err != nil {
		return nil, err
	}
	return normalizeDiscoveredModels(func() []string {
		ids := make([]string, 0, len(payload.Data))
		for _, item := range payload.Data {
			ids = append(ids, item.ID)
		}
		return ids
	}()), nil
}

func discoverGeminiModels(ctx context.Context, channel *model.AIProviderAccount) ([]DiscoveredModel, error) {
	endpoint := modelListEndpoint(channel.APIURL, "/v1beta/models")
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	query := parsed.Query()
	query.Set("key", channel.APIKey)
	parsed.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := doJSON(req, &payload); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(payload.Models))
	for _, item := range payload.Models {
		ids = append(ids, strings.TrimPrefix(item.Name, "models/"))
	}
	return normalizeDiscoveredModels(ids), nil
}

func modelListEndpoint(rawURL, suffix string) string {
	base := strings.TrimRight(rawURL, "/")
	if strings.HasSuffix(base, "/v1") {
		base = strings.TrimSuffix(base, "/v1")
	}
	if strings.HasSuffix(base, "/v1beta") {
		base = strings.TrimSuffix(base, "/v1beta")
	}
	if index := strings.Index(base, "/v1/"); index >= 0 {
		base = base[:index]
	}
	if index := strings.Index(base, "/v1beta/"); index >= 0 {
		base = base[:index]
	}
	return base + suffix
}

func doJSON(req *http.Request, target interface{}) error {
	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("模型发现请求失败: %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func normalizeDiscoveredModels(names []string) []DiscoveredModel {
	models := make([]DiscoveredModel, 0, len(names))
	seen := map[string]bool{}
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		modelType := discoveredType(name)
		if modelType == "" {
			continue
		}
		seen[name] = true
		models = append(models, DiscoveredModel{Name: name, Type: modelType})
	}
	return models
}

func discoveredType(name string) string {
	name = strings.ToLower(name)
	// Grok's image-to-video model contains both words. It must remain a video
	// task so it is routed to the asynchronous Grok video adapter.
	if strings.HasPrefix(name, "grok-") && strings.Contains(name, "video") {
		return "video"
	}
	if strings.Contains(name, "image") || strings.Contains(name, "dall-e") || strings.Contains(name, "seedream") || strings.Contains(name, "flux") {
		return "image"
	}
	if strings.Contains(name, "video") || strings.HasPrefix(name, "as-sd") || strings.Contains(name, "sora") || strings.Contains(name, "veo") || strings.Contains(name, "kling") || strings.Contains(name, "runway") {
		return "video"
	}
	if strings.Contains(name, "gpt") || strings.Contains(name, "gemini") || strings.Contains(name, "claude") || strings.Contains(name, "deepseek") || strings.Contains(name, "qwen") || strings.Contains(name, "glm") || strings.Contains(name, "grok") {
		return "text"
	}
	return ""
}

// JimengAdapter preserves the original channel contract. The international
// ZZ-based video API is intentionally isolated in JimengInternationalAdapter.
type JimengAdapter struct {
	apiConfig map[string]interface{}
}

func NewJimengAdapter(apiConfig map[string]interface{}) *JimengAdapter {
	return &JimengAdapter{apiConfig: apiConfig}
}

func (a *JimengAdapter) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, fmt.Errorf("即梦适配器尚未实现真实生成协议")
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
