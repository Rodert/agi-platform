package model

import (
	"time"

	"gorm.io/datatypes"
)

// AIModel AI模型配置
type AIModel struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name         string         `gorm:"size:100;not null" json:"name"`
	DisplayName  string         `gorm:"size:100;not null" json:"display_name"`
	Type         string         `gorm:"size:20;not null;index" json:"type"` // image/video
	Provider     string         `gorm:"size:50;not null" json:"provider"`
	ProviderAccountID *int64    `gorm:"index" json:"provider_account_id"`
	ProviderAccount   *AIProviderAccount `gorm:"foreignKey:ProviderAccountID" json:"provider_account,omitempty"`
	Description  string         `gorm:"size:500" json:"description"`
	LogoURL      string         `gorm:"size:255" json:"logo_url"`
	Tag          string         `gorm:"size:50" json:"tag"`
	Cost         int            `gorm:"not null" json:"cost"`
	APIConfig    datatypes.JSON `gorm:"type:json;not null" json:"api_config"`
	ParamsConfig datatypes.JSON `gorm:"type:json" json:"params_config"`
	IsActive     bool           `gorm:"default:true;index" json:"is_active"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
}

type AIProviderAccount struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"size:100;not null" json:"name"`
	Provider    string         `gorm:"size:50;not null;index" json:"provider"`
	APIURL      string         `gorm:"size:500;not null" json:"api_url"`
	APIKey      string         `gorm:"size:500;not null" json:"api_key"`
	ExtraConfig datatypes.JSON `gorm:"type:json" json:"extra_config"`
	IsActive    bool           `gorm:"default:true;index" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

func (AIProviderAccount) TableName() string { return "ai_provider_accounts" }

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
