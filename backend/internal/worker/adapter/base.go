package adapter

import (
	"context"
	"encoding/json"
	"time"
)

// GenerateRequest 生成请求
type GenerateRequest struct {
	ModelName string                 `json:"model_name"`
	Type      string                 `json:"type"`
	Prompt    string                 `json:"prompt"`
	Params    map[string]interface{} `json:"params"`
}

// GenerateResponse 生成响应
type GenerateResponse struct {
	ImageURL     string `json:"image_url"`
	ImageBase64  string `json:"image_base64"`
	VideoURL     string `json:"video_url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// Adapter AI 模型适配器接口
type Adapter interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}

// AsyncTaskAdapter is implemented by upstreams that accept a generation job
// and make its result available through a separate status endpoint. Worker
// orchestration only depends on this contract, never on a provider payload.
type AsyncTaskAdapter interface {
	Adapter
	Submit(ctx context.Context, req *GenerateRequest) (*AsyncTask, error)
	Poll(ctx context.Context, providerTaskID string) (*AsyncTask, error)
}

// AsyncTask is the normalized state of an upstream generation task.
type AsyncTask struct {
	ProviderTaskID string            `json:"provider_task_id"`
	Status         string            `json:"status"` // queued/processing/succeeded/failed
	Progress       int               `json:"progress"`
	ErrorMessage   string            `json:"error_message"`
	Result         *GenerateResponse `json:"result,omitempty"`
	RawResponse    json.RawMessage   `json:"raw_response,omitempty"`
}

// PollingConfig is optional. Providers may tune their own API-specific polling
// cadence while the worker still owns persistence and result storage.
type PollingConfig interface {
	PollInterval() time.Duration
	PollTimeout() time.Duration
}

type Factory func(config map[string]interface{}) (Adapter, error)

// DiscoveredModel is the normalized result returned by a channel adapter.
// Capability schemas are intentionally not stored here; they are owned by the
// global model catalog and shared by every channel that supports the model.
type DiscoveredModel struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
