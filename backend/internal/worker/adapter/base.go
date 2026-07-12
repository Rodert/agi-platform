package adapter

import (
	"context"
)

// GenerateRequest 生成请求
type GenerateRequest struct {
	Prompt string                 `json:"prompt"`
	Params map[string]interface{} `json:"params"`
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
