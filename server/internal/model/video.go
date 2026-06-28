package model

import (
	"time"

	"gorm.io/datatypes"
)

const (
	VideoTaskStatusPending   = "pending"
	VideoTaskStatusRunning   = "running"
	VideoTaskStatusSucceeded = "succeeded"
	VideoTaskStatusFailed    = "failed"
	VideoTaskStatusTimeout   = "timeout"
)

type VideoModel struct {
	BaseModel
	Code                  string         `gorm:"column:code" json:"code"`
	DisplayName           string         `gorm:"column:display_name" json:"display_name"`
	Description           string         `gorm:"column:description" json:"description"`
	PriceCredits          int64          `gorm:"column:price_credits" json:"price_credits"`
	SupportedAspectRatios datatypes.JSON `gorm:"column:supported_aspect_ratios" json:"supported_aspect_ratios"`
	SupportedSeconds      datatypes.JSON `gorm:"column:supported_seconds" json:"supported_seconds"`
	Enabled               bool           `gorm:"column:enabled" json:"enabled"`
	Recommended           bool           `gorm:"column:recommended" json:"recommended"`
	SortOrder             int            `gorm:"column:sort_order" json:"sort_order"`
}

func (VideoModel) TableName() string {
	return "video_models"
}

type VideoModelRoute struct {
	ID                uint64         `gorm:"primaryKey;column:id" json:"id"`
	ModelID           uint64         `gorm:"column:model_id" json:"model_id"`
	ProviderID        uint64         `gorm:"column:provider_id" json:"provider_id"`
	ProviderKeyID     *uint64        `gorm:"column:provider_key_id" json:"provider_key_id,omitempty"`
	ProviderModelName string         `gorm:"column:provider_model_name" json:"provider_model_name"`
	Enabled           bool           `gorm:"column:enabled" json:"enabled"`
	Priority          int            `gorm:"column:priority" json:"priority"`
	Weight            int            `gorm:"column:weight" json:"weight"`
	ExtraConfig       datatypes.JSON `gorm:"column:extra_config" json:"extra_config"`
	CreatedAt         time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (VideoModelRoute) TableName() string {
	return "video_model_routes"
}

type VideoTask struct {
	ID               uint64         `gorm:"primaryKey;column:id" json:"id"`
	TaskNo           string         `gorm:"column:task_no" json:"task_no"`
	UserID           uint64         `gorm:"column:user_id" json:"user_id"`
	APIKeyID         *uint64        `gorm:"column:api_key_id" json:"api_key_id,omitempty"`
	ModelID          uint64         `gorm:"column:model_id" json:"model_id"`
	RouteID          *uint64        `gorm:"column:route_id" json:"route_id,omitempty"`
	ProviderID       *uint64        `gorm:"column:provider_id" json:"provider_id,omitempty"`
	ProviderKeyID    *uint64        `gorm:"column:provider_key_id" json:"provider_key_id,omitempty"`
	Source           string         `gorm:"column:source" json:"source"`
	Prompt           string         `gorm:"column:prompt" json:"prompt"`
	Seconds          int            `gorm:"column:seconds" json:"seconds"`
	AspectRatio      string         `gorm:"column:aspect_ratio" json:"aspect_ratio"`
	Images           datatypes.JSON `gorm:"column:images" json:"images"`
	Videos           datatypes.JSON `gorm:"column:videos" json:"videos"`
	Audios           datatypes.JSON `gorm:"column:audios" json:"audios"`
	Status           string         `gorm:"column:status" json:"status"`
	Progress         int            `gorm:"column:progress" json:"progress"`
	CreditsUsed      int64          `gorm:"column:credits_used" json:"credits_used"`
	RefundStatus     string         `gorm:"column:refund_status" json:"refund_status"`
	ProviderTaskID   string         `gorm:"column:provider_task_id" json:"provider_task_id"`
	ProviderResponse datatypes.JSON `gorm:"column:provider_response" json:"provider_response"`
	ErrorMessage     string         `gorm:"column:error_message" json:"error_message"`
	StartedAt        *time.Time     `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt      *time.Time     `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt        time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at" json:"updated_at"`
}

func (VideoTask) TableName() string {
	return "video_tasks"
}

type VideoAsset struct {
	BaseModel
	TaskID          uint64  `gorm:"column:task_id" json:"task_id"`
	UserID          uint64  `gorm:"column:user_id" json:"user_id"`
	ModelID         uint64  `gorm:"column:model_id" json:"model_id"`
	URL             string  `gorm:"column:url" json:"url"`
	StorageProvider string  `gorm:"column:storage_provider" json:"storage_provider"`
	StorageBucket   string  `gorm:"column:storage_bucket" json:"storage_bucket"`
	StorageKey      string  `gorm:"column:storage_key" json:"storage_key"`
	MimeType        string  `gorm:"column:mime_type" json:"mime_type"`
	SizeBytes       *int64  `gorm:"column:size_bytes" json:"size_bytes,omitempty"`
	DurationSeconds *int    `gorm:"column:duration_seconds" json:"duration_seconds,omitempty"`
	Prompt          *string `gorm:"column:prompt" json:"prompt,omitempty"`
	Status          string  `gorm:"column:status" json:"status"`
	IsPublic        bool    `gorm:"column:is_public" json:"is_public"`
	DownloadCount   int64   `gorm:"column:download_count" json:"download_count"`
}

func (VideoAsset) TableName() string {
	return "video_assets"
}
