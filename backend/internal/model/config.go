package model

import (
	"time"

	"gorm.io/datatypes"
)

// AIModel AI模型配置
type AIModel struct {
	ID          int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `gorm:"size:100;uniqueIndex;not null" json:"name"`
	DisplayName string `gorm:"size:100;not null" json:"display_name"`
	Type        string `gorm:"size:20;not null;index" json:"type"` // image/video/text
	// Provider is retained for historical rows only. Routing is defined by ChannelModels.
	Provider      string         `gorm:"size:50;not null" json:"provider"`
	Description   string         `gorm:"size:500" json:"description"`
	LogoURL       string         `gorm:"size:255" json:"logo_url"`
	Tag           string         `gorm:"size:50" json:"tag"`
	Cost          int            `gorm:"not null" json:"cost"`
	APIConfig     datatypes.JSON `gorm:"type:json;not null" json:"api_config"`
	ParamsConfig  datatypes.JSON `gorm:"type:json" json:"params_config"`
	IsActive      bool           `gorm:"default:true;index" json:"is_active"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
	ChannelModels []ChannelModel `gorm:"foreignKey:ModelID" json:"channel_models,omitempty"`
}

type AIProviderAccount struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name          string         `gorm:"size:100;not null" json:"name"`
	Provider      string         `gorm:"size:50;not null;index" json:"provider"`
	APIURL        string         `gorm:"size:500;not null" json:"api_url"`
	APIKey        string         `gorm:"size:500;not null" json:"api_key"`
	ExtraConfig   datatypes.JSON `gorm:"type:json" json:"extra_config"`
	IsActive      bool           `gorm:"default:true;index" json:"is_active"`
	Priority      int            `gorm:"default:100;index" json:"priority"`
	HealthStatus  string         `gorm:"size:20;default:unknown" json:"health_status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	ChannelModels []ChannelModel `gorm:"foreignKey:ChannelID" json:"channel_models,omitempty"`
}

func (AIProviderAccount) TableName() string { return "ai_provider_accounts" }

// ChannelModel only describes whether a channel account can call a global model.
// Model capabilities live exclusively on AIModel, so every channel uses the same schema.
type ChannelModel struct {
	ID        int64              `gorm:"primaryKey;autoIncrement" json:"id"`
	ChannelID int64              `gorm:"not null;uniqueIndex:idx_channel_model" json:"channel_id"`
	ModelID   int64              `gorm:"not null;uniqueIndex:idx_channel_model;index" json:"model_id"`
	IsActive  bool               `gorm:"default:true;index" json:"is_active"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	Channel   *AIProviderAccount `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Model     *AIModel           `gorm:"foreignKey:ModelID" json:"model,omitempty"`
}

func (ChannelModel) TableName() string { return "channel_models" }

func (AIModel) TableName() string {
	return "ai_models"
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Key         string    `gorm:"size:100;uniqueIndex;not null" json:"key"`
	Value       string    `gorm:"type:text;not null" json:"value"`
	Type        string    `gorm:"size:20;not null" json:"type"` // string/int/json/bool
	Category    string    `gorm:"size:50;index" json:"category"`
	Description string    `gorm:"size:255" json:"description"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// TaskConfig contains execution policy shared by all image and video tasks.
// It is intentionally a dedicated singleton instead of generic key/value data,
// so task processing can evolve without coupling to unrelated system settings.
type TaskConfig struct {
	ID               int64     `gorm:"primaryKey;default:1" json:"id"`
	MaxActiveTasks   int       `gorm:"not null;default:50" json:"max_active_tasks"`
	PromptMaxLength  int       `gorm:"not null;default:5000" json:"prompt_max_length"`
	MaxRetryAttempts int       `gorm:"not null;default:0" json:"max_retry_attempts"`
	UpdatedAt        time.Time `gorm:"not null" json:"updated_at"`
}

func (TaskConfig) TableName() string {
	return "task_configs"
}

// PromptOptimizationConfig is the singleton policy for AI prompt refinement.
// It intentionally remains separate from task execution settings because it is
// a synchronous text operation and does not create a generation task.
type PromptOptimizationConfig struct {
	ID                 int64     `gorm:"primaryKey;default:1" json:"id"`
	IsActive           bool      `gorm:"not null;default:false" json:"is_active"`
	ModelName          string    `gorm:"size:100" json:"model_name"`
	SystemPrompt       string    `gorm:"type:text;not null" json:"system_prompt"`
	MaxInputLength     int       `gorm:"not null;default:5000" json:"max_input_length"`
	CreditCost         int       `gorm:"not null;default:0" json:"credit_cost"`
	RateLimitPerMinute int       `gorm:"not null;default:5" json:"rate_limit_per_minute"`
	UpdatedAt          time.Time `gorm:"not null" json:"updated_at"`
}

func (PromptOptimizationConfig) TableName() string {
	return "prompt_optimization_configs"
}

// PromptOptimizationLog preserves request, billing, routing, and provider
// errors without coupling prompt improvement to the generation task lifecycle.
type PromptOptimizationLog struct {
	ID              int64              `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          int64              `gorm:"not null;index:idx_prompt_optimization_user_created" json:"user_id"`
	ModelName       string             `gorm:"size:100;not null" json:"model_name"`
	ChannelID       int64              `gorm:"not null;index" json:"channel_id"`
	TargetType      string             `gorm:"size:20;not null" json:"target_type"`
	TargetModelName string             `gorm:"size:100" json:"target_model_name"`
	Params          datatypes.JSON     `gorm:"type:json" json:"params"`
	OriginalPrompt  string             `gorm:"type:text;not null" json:"original_prompt"`
	OptimizedPrompt string             `gorm:"type:text" json:"optimized_prompt"`
	CreditCost      int                `gorm:"not null;default:0" json:"credit_cost"`
	Status          string             `gorm:"size:20;not null;index" json:"status"`
	ErrorMsg        string             `gorm:"size:1000" json:"error_msg"`
	LatencyMS       int                `gorm:"not null;default:0" json:"latency_ms"`
	CreatedAt       time.Time          `gorm:"not null;index:idx_prompt_optimization_user_created" json:"created_at"`
	UpdatedAt       time.Time          `gorm:"not null" json:"updated_at"`
	User            *User              `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Channel         *AIProviderAccount `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
}

func (PromptOptimizationLog) TableName() string {
	return "prompt_optimization_logs"
}

// Category 分类配置
type Category struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Type      string    `gorm:"size:30;not null;index" json:"type"` // work_category/aspect_ratio
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (Category) TableName() string {
	return "categories"
}

// EmailConfig 邮箱配置（单条记录）
type EmailConfig struct {
	ID           int64     `gorm:"primaryKey;default:1" json:"id"`
	SMTPHost     string    `gorm:"size:100;not null" json:"smtp_host"`
	SMTPPort     int       `gorm:"not null" json:"smtp_port"`
	SMTPUser     string    `gorm:"size:100;not null" json:"smtp_user"`
	SMTPPassword string    `gorm:"size:255;not null" json:"smtp_password"`
	SMTPSSL      bool      `gorm:"default:false" json:"smtp_ssl"`
	FromName     string    `gorm:"size:100;not null" json:"from_name"`
	FromEmail    string    `gorm:"size:100;not null" json:"from_email"`
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

func (EmailConfig) TableName() string {
	return "email_config"
}
