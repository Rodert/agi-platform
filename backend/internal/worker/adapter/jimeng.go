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

// JimengInternationalAdapter implements ZZ API's international Jimeng video endpoints.
// The provider returns an asynchronous task and keeps completed video content
// behind an authenticated download endpoint.
type JimengInternationalAdapter struct {
	config map[string]interface{}
	client *http.Client
}

func NewJimengInternationalAdapter(config map[string]interface{}) *JimengInternationalAdapter {
	return &JimengInternationalAdapter{config: config, client: &http.Client{Timeout: 45 * time.Second}}
}

func (a *JimengInternationalAdapter) Generate(context.Context, *GenerateRequest) (*GenerateResponse, error) {
	return nil, fmt.Errorf("即梦国际视频任务必须通过异步任务接口执行")
}

func (a *JimengInternationalAdapter) Submit(ctx context.Context, request *GenerateRequest) (*AsyncTask, error) {
	if request.Type != "video" {
		return nil, fmt.Errorf("即梦国际渠道仅支持视频生成")
	}
	payload, err := json.Marshal(a.createPayload(request))
	if err != nil {
		return nil, fmt.Errorf("编码即梦国际创建任务请求失败: %w", err)
	}
	raw, err := a.do(ctx, http.MethodPost, a.endpoint("create_path", "/videos", ""), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	return a.parseTask(raw, "")
}

func (a *JimengInternationalAdapter) Poll(ctx context.Context, providerTaskID string) (*AsyncTask, error) {
	if strings.TrimSpace(providerTaskID) == "" {
		return nil, fmt.Errorf("即梦国际上游任务 ID 不能为空")
	}
	raw, err := a.do(ctx, http.MethodGet, a.endpoint("status_path", "/videos/{task_id}", providerTaskID), nil)
	if err != nil {
		return nil, err
	}
	return a.parseTask(raw, providerTaskID)
}

func (a *JimengInternationalAdapter) PollInterval() time.Duration {
	return time.Duration(configInt(a.config, "poll_interval_seconds", 5, 1, 60)) * time.Second
}

func (a *JimengInternationalAdapter) PollTimeout() time.Duration {
	return time.Duration(configInt(a.config, "poll_timeout_seconds", 3600, 30, 3600)) * time.Second
}

func (a *JimengInternationalAdapter) createPayload(request *GenerateRequest) map[string]interface{} {
	body := map[string]interface{}{"model": request.ModelName, "prompt": request.Prompt}
	images := make([]string, 0, 4)
	for key, value := range request.Params {
		switch key {
		case "reference_image_url", "first_frame_url", "last_frame_url":
			if image, ok := value.(string); ok && strings.TrimSpace(image) != "" {
				images = append(images, image)
			}
		case "reference_image_urls", "images":
			images = append(images, stringValues(value)...)
		case "ratio":
			body["aspect_ratio"] = value
		case "duration":
			if seconds := normalizeSeconds(value); seconds != "" {
				body["seconds"] = seconds
			}
		}
	}
	if len(images) > 0 {
		body["images"] = images
	}
	return body
}

func (a *JimengInternationalAdapter) parseTask(raw []byte, fallbackID string) (*AsyncTask, error) {
	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		if strings.HasPrefix(strings.TrimSpace(string(raw)), "<") {
			return nil, fmt.Errorf("即梦国际上游返回了网页而非 JSON，请检查渠道 API 地址；ZZ API 应配置为 https://zz1cc.cc.cd/v1")
		}
		return nil, fmt.Errorf("解析即梦国际任务响应失败: %w", err)
	}
	root, _ := payload.(map[string]interface{})
	data := nestedMap(root, "data")
	candidates := []map[string]interface{}{data, nestedMap(data, "task"), nestedMap(data, "data"), root}
	taskID := fallbackID
	if taskID == "" {
		for _, candidate := range candidates {
			if id := firstString(candidate, "task_id", "id", "video_id"); id != "" {
				taskID = id
				break
			}
		}
	}
	if taskID == "" {
		return nil, fmt.Errorf("即梦国际上游未返回任务 ID")
	}

	status := ""
	progress := 0
	for _, candidate := range candidates {
		if status == "" {
			status = normalizeJimengStatus(firstString(candidate, "status", "state"))
		}
		if progress == 0 {
			progress = boundedProgress(firstNumber(candidate, "progress", "percentage"))
		}
	}
	if status == "" {
		status = "processing"
	}
	task := &AsyncTask{ProviderTaskID: taskID, Status: status, Progress: progress, RawResponse: append(json.RawMessage(nil), raw...)}
	if status == "failed" {
		task.ErrorMessage = jimengErrorMessage(candidates)
		return task, nil
	}
	if status == "succeeded" {
		for _, candidate := range candidates {
			if videoURL := firstString(candidate, "video_url", "url", "output_url", "download_url"); videoURL != "" {
				task.Result = &GenerateResponse{VideoURL: videoURL}
				return task, nil
			}
		}
		task.Result = &GenerateResponse{
			VideoURL:     a.endpoint("content_path", "/videos/{task_id}/content", taskID),
			VideoHeaders: map[string]string{"Authorization": a.authorizationHeader()},
		}
	}
	return task, nil
}

func (a *JimengInternationalAdapter) endpoint(pathKey, defaultPath, taskID string) string {
	path := configString(a.config, pathKey, defaultPath)
	path = strings.ReplaceAll(path, "{task_id}", url.PathEscape(taskID))
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	base := strings.TrimRight(configString(a.config, "api_url", ""), "/")
	if base == "" {
		return ""
	}
	// ZZ API documentation uses a /v1 base URL while its model endpoint is
	// written as /v1/models. Accept both forms without producing /v1/v1/models.
	if strings.HasSuffix(base, "/v1") && strings.HasPrefix(path, "/v1/") {
		base = strings.TrimSuffix(base, "/v1")
	}
	return base + "/" + strings.TrimLeft(path, "/")
}

func (a *JimengInternationalAdapter) authorizationHeader() string {
	return "Bearer " + configString(a.config, "api_key", "")
}

func (a *JimengInternationalAdapter) do(ctx context.Context, method, endpoint string, body io.Reader) ([]byte, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("即梦国际渠道未配置 API 地址")
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", a.authorizationHeader())
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 10*1024*1024))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("即梦国际上游请求失败: %s: %s", resp.Status, strings.TrimSpace(string(raw)))
	}
	return raw, nil
}

func discoverJimengInternationalModels(ctx context.Context, channel *model.AIProviderAccount) ([]DiscoveredModel, error) {
	config := map[string]interface{}{"api_url": channel.APIURL, "api_key": channel.APIKey}
	if len(channel.ExtraConfig) > 0 {
		_ = json.Unmarshal(channel.ExtraConfig, &config)
		config["api_url"], config["api_key"] = channel.APIURL, channel.APIKey
	}
	adapter := NewJimengInternationalAdapter(config)
	raw, err := adapter.do(ctx, http.MethodGet, adapter.endpoint("models_path", "/v1/models", ""), nil)
	if err != nil {
		return nil, err
	}
	var payload interface{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("解析即梦国际模型列表失败: %w", err)
	}
	return normalizeDiscoveredModels(modelNames(payload)), nil
}

func normalizeJimengStatus(value string) string {
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

func jimengErrorMessage(candidates []map[string]interface{}) string {
	for _, candidate := range candidates {
		if message := firstString(candidate, "error_message", "message", "fail_reason"); message != "" {
			return message
		}
		if message := nestedString(candidate, "error", "message"); message != "" {
			return message
		}
	}
	return "上游任务失败"
}

func normalizeSeconds(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSuffix(strings.TrimSpace(typed), "s")
	case float64:
		return strconv.FormatInt(int64(typed), 10)
	case int:
		return strconv.Itoa(typed)
	default:
		return ""
	}
}
