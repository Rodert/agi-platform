package model

import (
	"time"

	"gorm.io/datatypes"
)

// GenerationRequest 生成请求模型
type GenerationRequest struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64          `gorm:"not null;index" json:"user_id"`
	Prompt    string         `gorm:"type:text;not null" json:"prompt"`
	ModelName string         `gorm:"size:50;not null" json:"model_name"`
	Type      string         `gorm:"size:20;not null" json:"type"` // image/video/product
	Params    datatypes.JSON `gorm:"type:json" json:"params"`
	Cost      int            `gorm:"not null" json:"cost"`
	TaskID    int64          `gorm:"index" json:"task_id"`
	CreatedAt time.Time      `gorm:"not null;index" json:"created_at"`
}

func (GenerationRequest) TableName() string {
	return "generation_requests"
}

// Task 任务模型
type Task struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64      `gorm:"not null;index:idx_user_status" json:"user_id"`
	RequestID    int64      `gorm:"not null;index" json:"request_id"`
	Title        string     `gorm:"size:255;not null" json:"title"`
	Type         string     `gorm:"size:20;not null" json:"type"` // image/video/product
	Status       string     `gorm:"size:20;not null;default:queued;index:idx_user_status" json:"status"` // queued/processing/success/failed
	Progress     int        `gorm:"default:0" json:"progress"`
	Prompt       string     `gorm:"type:text;not null" json:"prompt"`
	ModelName    string     `gorm:"size:50;not null" json:"model_name"`
	ResultURL    string     `gorm:"size:500" json:"result_url"`
	ThumbnailURL string     `gorm:"size:500" json:"thumbnail_url"`
	ErrorMsg     string     `gorm:"size:500" json:"error_msg"`
	Cost         int        `gorm:"not null" json:"cost"`
	CreatedAt    time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"not null" json:"updated_at"`
	CompletedAt  *time.Time `gorm:"index" json:"completed_at"`
}

func (Task) TableName() string {
	return "tasks"
}
