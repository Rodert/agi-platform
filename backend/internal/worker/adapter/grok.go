package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
)

// GrokAdapter isolates Grok's asynchronous video API from task orchestration.
// The documented paths can be overridden in channel.extra_config for compatible
// gateways: models_path, create_path, status_path, reference_field,
// poll_interval_seconds and poll_timeout_seconds.
type GrokAdapter struct {
	config map[string]interface{}
	client *http.Client
}

func NewGrokAdapter(config map[string]interface{}) *GrokAdapter {
	return &GrokAdapter{config: config, client: &http.Client{Timeout: 45 * time.Second}}
}

// Generate is deliberately not used for Grok. Keeping the base Adapter method
// lets the registry route synchronous and asynchronous providers uniformly.
func (a *GrokAdapter) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, fmt.Errorf("Grok 视频任务必须通过异步任务接口执行")
}

func (a *GrokAdapter) Submit(ctx context.Context, req *GenerateRequest) (*AsyncTask, error) {
	if req.Type != "video" {
		return nil, fmt.Errorf("Grok 渠道仅支持视频生成")
	}
	payload, err := json.Marshal(a.createPayload(req))
	if err != nil {
		return nil, fmt.Errorf("编码 Grok 创建任务请求失败: %w", err)
	}
	response, err := a.do(ctx, http.MethodPost, a.endpoint("create_path", "/v1/video/generations", ""), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	return parseGrokTask(response)
}

func (a *GrokAdapter) Poll(ctx context.Context, providerTaskID string) (*AsyncTask, error) {
	if strings.TrimSpace(providerTaskID) == "" {
		return nil, fmt.Errorf("Grok 上游任务 ID 不能为空")
	}
	response, err := a.do(ctx, http.MethodGet, a.endpoint("status_path", "/v1/video/generations/{task_id}", providerTaskID), nil)
	if err != nil {
		return nil, err
	}
	return parseGrokTask(response)
}

func (a *GrokAdapter) PollInterval() time.Duration {
	return time.Duration(configInt(a.config, "poll_interval_seconds", 5, 1, 60)) * time.Second
}

func (a *GrokAdapter) PollTimeout() time.Duration {
	return time.Duration(configInt(a.config, "poll_timeout_seconds", 900, 30, 3600)) * time.Second
}

func (a *GrokAdapter) createPayload(req *GenerateRequest) map[string]interface{} {
	body := map[string]interface{}{"model": req.ModelName, "prompt": req.Prompt}
	references := make([]string, 0, 3)
	for key, value := range req.Params {
		switch key {
		case "reference_image_url", "first_frame_url", "last_frame_url":
			if raw, ok := value.(string); ok && strings.TrimSpace(raw) != "" {
				references = append(references, raw)
			}
		case "reference_image_urls", "image_urls":
			references = append(references, stringValues(value)...)
		case "ratio":
			body["aspect_ratio"] = value
		case "duration":
			body["seconds"] = value
		default:
			body[key] = value
		}
	}
	if len(references) > 0 {
		field := configString(a.config, "reference_field", "images")
		if len(references) == 1 && field == "image_url" {
			body[field] = references[0]
		} else {
			body[field] = references
		}
	}
	return body
}

func (a *GrokAdapter) endpoint(pathKey, defaultPath, taskID string) string {
	path := configString(a.config, pathKey, defaultPath)
	path = strings.ReplaceAll(path, "{task_id}", url.PathEscape(taskID))
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	base := strings.TrimRight(configString(a.config, "api_url", ""), "/")
	if base == "" {
		return ""
	}
	if path == "" {
		return base
	}
	if strings.HasSuffix(base, strings.TrimRight(path, "/")) {
		return base
	}
	return base + "/" + strings.TrimLeft(path, "/")
}

func (a *GrokAdapter) do(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("Grok 渠道未配置 API 地址")
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	key := configString(a.config, "api_key", "")
	if key != "" {
		header := configString(a.config, "api_key_header", "Authorization")
		prefix := configString(a.config, "api_key_prefix", "Bearer ")
		req.Header.Set(header, prefix+key)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("Grok 上游请求失败: %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	return data, nil
}

func discoverGrokModels(ctx context.Context, channel *model.AIProviderAccount) ([]DiscoveredModel, error) {
	config := map[string]interface{}{"api_url": channel.APIURL, "api_key": channel.APIKey}
	if len(channel.ExtraConfig) > 0 {
		_ = json.Unmarshal(channel.ExtraConfig, &config)
		config["api_url"], config["api_key"] = channel.APIURL, channel.APIKey
	}
	adapter := NewGrokAdapter(config)
	response, err := adapter.do(ctx, http.MethodGet, adapter.endpoint("models_path", "/v1/models", ""), nil)
	if err != nil {
		return nil, err
	}
	var payload interface{}
	if err := json.Unmarshal(response, &payload); err != nil {
		return nil, fmt.Errorf("解析 Grok 模型列表失败: %w", err)
	}
	names := modelNames(payload)
	return normalizeDiscoveredModels(names), nil
}

func parseGrokTask(raw []byte) (*AsyncTask, error) {
	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("解析 Grok 任务响应失败: %w", err)
	}
	root, _ := payload.(map[string]interface{})
	data := nestedMap(root, "data")
	hasDataEnvelope := data != nil
	if data == nil {
		data = root
	}
	task := &AsyncTask{
		ProviderTaskID: firstString(data, "task_id", "id", "video_id"),
		Status:         normalizeGrokStatus(firstString(data, "status", "state")),
		Progress:       boundedProgress(firstNumber(data, "progress", "percentage")),
		ErrorMessage:   firstString(data, "error_message", "error", "message"),
		RawResponse:    append(json.RawMessage(nil), raw...),
	}
	if task.ProviderTaskID == "" && hasDataEnvelope {
		task.ProviderTaskID = firstString(root, "task_id", "id")
	}
	if task.Status == "" {
		task.Status = "processing"
	}
	if task.Status == "failed" && task.ErrorMessage == "" {
		task.ErrorMessage = nestedString(root, "error", "message")
	}
	if task.Status == "succeeded" {
		result := map[string]interface{}{}
		for _, candidate := range []map[string]interface{}{data, nestedMap(data, "result"), nestedMap(data, "output"), nestedMap(root, "data"), root} {
			if candidate == nil {
				continue
			}
			if resultString(result, "video") == "" {
				result["video"] = firstString(candidate, "video_url", "url", "output_url", "download_url")
			}
			if resultString(result, "thumbnail") == "" {
				result["thumbnail"] = firstString(candidate, "thumbnail_url", "cover_url", "poster_url")
			}
		}
		videoURL := resultString(result, "video")
		if videoURL == "" {
			return nil, fmt.Errorf("Grok 任务已完成但未返回视频地址")
		}
		thumbnailURL := resultString(result, "thumbnail")
		task.Result = &GenerateResponse{VideoURL: videoURL, ThumbnailURL: thumbnailURL}
	}
	return task, nil
}

func resultString(value map[string]interface{}, key string) string {
	result, _ := value[key].(string)
	return result
}

func normalizeGrokStatus(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "queued", "pending", "created", "submitted":
		return "queued"
	case "processing", "running", "in_progress", "generating":
		return "processing"
	case "success", "succeeded", "completed", "done":
		return "succeeded"
	case "failed", "failure", "error", "cancelled", "canceled", "expired":
		return "failed"
	default:
		return ""
	}
}

func modelNames(payload interface{}) []string {
	root, _ := payload.(map[string]interface{})
	items, _ := root["data"].([]interface{})
	if len(items) == 0 {
		items, _ = root["models"].([]interface{})
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		if row, ok := item.(map[string]interface{}); ok {
			if name := firstString(row, "id", "name", "model"); name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

func nestedMap(value map[string]interface{}, key string) map[string]interface{} {
	if value == nil {
		return nil
	}
	item, _ := value[key].(map[string]interface{})
	return item
}

func nestedString(value map[string]interface{}, parent, key string) string {
	return firstString(nestedMap(value, parent), key)
}

func firstString(value map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if raw, ok := value[key]; ok {
			switch typed := raw.(type) {
			case string:
				if strings.TrimSpace(typed) != "" {
					return typed
				}
			case json.Number:
				return typed.String()
			case float64:
				return strconv.FormatInt(int64(typed), 10)
			}
		}
	}
	return ""
}

func firstNumber(value map[string]interface{}, keys ...string) int {
	for _, key := range keys {
		if raw, ok := value[key]; ok {
			switch typed := raw.(type) {
			case float64:
				return int(typed)
			case string:
				if number, err := strconv.Atoi(typed); err == nil {
					return number
				}
			}
		}
	}
	return 0
}

func boundedProgress(value int) int {
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func stringValues(value interface{}) []string {
	items, ok := value.([]interface{})
	if !ok {
		if strings, ok := value.([]string); ok {
			return strings
		}
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		if value, ok := item.(string); ok && strings.TrimSpace(value) != "" {
			values = append(values, value)
		}
	}
	return values
}

func configString(config map[string]interface{}, key, fallback string) string {
	if value, ok := config[key].(string); ok && strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func configInt(config map[string]interface{}, key string, fallback, min, max int) int {
	value := fallback
	switch raw := config[key].(type) {
	case float64:
		value = int(raw)
	case int:
		value = raw
	case string:
		if parsed, err := strconv.Atoi(raw); err == nil {
			value = parsed
		}
	}
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
