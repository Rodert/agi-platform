package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	ImageTaskStatusPending   = "pending"
	ImageTaskStatusRunning   = "running"
	ImageTaskStatusSucceeded = "succeeded"
	ImageTaskStatusFailed    = "failed"
)

type ImageTask struct {
	ID                uint64         `gorm:"primaryKey;column:id" json:"id"`
	TaskNo            string         `gorm:"column:task_no" json:"task_no"`
	UserID            uint64         `gorm:"column:user_id" json:"user_id"`
	APIKeyID          *uint64        `gorm:"column:api_key_id" json:"api_key_id,omitempty"`
	ModelID           uint64         `gorm:"column:model_id" json:"model_id"`
	RouteID           *uint64        `gorm:"column:route_id" json:"route_id,omitempty"`
	ProviderID        *uint64        `gorm:"column:provider_id" json:"provider_id,omitempty"`
	ProviderKeyID     *uint64        `gorm:"column:provider_key_id" json:"provider_key_id,omitempty"`
	Source            string         `gorm:"column:source" json:"source"`
	Prompt            string         `gorm:"column:prompt" json:"prompt"`
	NegativePrompt    *string        `gorm:"column:negative_prompt" json:"negative_prompt,omitempty"`
	Size              string         `gorm:"column:size" json:"size"`
	NumImages         int            `gorm:"column:num_images" json:"num_images"`
	Status            string         `gorm:"column:status" json:"status"`
	Progress          int            `gorm:"column:progress" json:"progress"`
	CreditsUsed       int64          `gorm:"column:credits_used" json:"credits_used"`
	RefundStatus      string         `gorm:"column:refund_status" json:"refund_status"`
	ProviderRequestID string         `gorm:"column:provider_request_id" json:"provider_request_id"`
	ProviderResponse  datatypes.JSON `gorm:"column:provider_response" json:"provider_response"`
	ErrorCode         string         `gorm:"column:error_code" json:"error_code"`
	ErrorMessage      string         `gorm:"column:error_message" json:"error_message"`
	StartedAt         *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt       *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (ImageTask) TableName() string {
	return "image_tasks"
}

type ImageAsset struct {
	BaseModel
	TaskID          uint64  `gorm:"column:task_id" json:"task_id"`
	UserID          uint64  `gorm:"column:user_id" json:"user_id"`
	ModelID         uint64  `gorm:"column:model_id" json:"model_id"`
	URL             string  `gorm:"column:url" json:"url"`
	StorageProvider string  `gorm:"column:storage_provider" json:"storage_provider"`
	StorageBucket   string  `gorm:"column:storage_bucket" json:"storage_bucket"`
	StorageKey      string  `gorm:"column:storage_key" json:"storage_key"`
	Width           *int    `gorm:"column:width" json:"width,omitempty"`
	Height          *int    `gorm:"column:height" json:"height,omitempty"`
	MimeType        string  `gorm:"column:mime_type" json:"mime_type"`
	SizeBytes       *int64  `gorm:"column:size_bytes" json:"size_bytes,omitempty"`
	Prompt          *string `gorm:"column:prompt" json:"prompt,omitempty"`
	Status          string  `gorm:"column:status" json:"status"`
	IsPublic        bool    `gorm:"column:is_public" json:"is_public"`
	DownloadCount   int64   `gorm:"column:download_count" json:"download_count"`
	ViolationStatus string  `gorm:"column:violation_status" json:"violation_status"`
}

func (ImageAsset) TableName() string {
	return "image_assets"
}
