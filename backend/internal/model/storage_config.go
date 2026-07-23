package model

import "time"

// StorageConfig 存储配置
type StorageConfig struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`            // 配置名称
	Type      string    `gorm:"size:50;not null" json:"type"`             // 存储类型: local, tencent_cos, aliyun_oss, cloudflare
	LocalPath string    `gorm:"size:255" json:"local_path"`               // 本地存储路径
	Endpoint  string    `gorm:"size:255" json:"endpoint"`                 // 端点
	AccessKey string    `gorm:"size:255" json:"access_key"`               // AccessKey
	SecretKey string    `gorm:"size:255" json:"secret_key"`               // SecretKey
	Bucket    string    `gorm:"size:100" json:"bucket"`                   // 桶名称
	Region    string    `gorm:"size:50" json:"region"`                    // 区域
	Domain    string    `gorm:"size:255" json:"domain"`                   // CDN 域名
	IsEnabled bool      `gorm:"default:false;not null" json:"is_enabled"` // 是否启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (StorageConfig) TableName() string {
	return "storage_configs"
}

// ResourcePolicy controls how each generated or uploaded resource is stored.
type ResourcePolicy struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	ResourceType  string    `gorm:"size:30;uniqueIndex;not null" json:"resource_type"`
	KeyPrefix     string    `gorm:"size:100;not null" json:"key_prefix"`
	RetentionDays int       `gorm:"not null;default:0" json:"retention_days"`
	IsPublic      bool      `gorm:"not null;default:true" json:"is_public"`
	CacheMaxAge   int       `gorm:"not null;default:86400" json:"cache_max_age"`
	MaxSizeMB     int       `gorm:"not null" json:"max_size_mb"`
	UpdatedAt     time.Time `gorm:"not null" json:"updated_at"`
}

func (ResourcePolicy) TableName() string { return "resource_policies" }

// MediaAsset is storage-provider-neutral metadata for a generated or uploaded object.
type MediaAsset struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID          *int64     `gorm:"index" json:"task_id"`
	UserID          int64      `gorm:"not null;index" json:"user_id"`
	StorageConfigID int64      `gorm:"not null;index" json:"storage_config_id"`
	ResourceType    string     `gorm:"size:30;not null;index" json:"resource_type"`
	ObjectKey       string     `gorm:"size:500;not null;uniqueIndex" json:"object_key"`
	PublicURL       string     `gorm:"size:1000" json:"public_url"`
	ContentType     string     `gorm:"size:100;not null" json:"content_type"`
	SizeBytes       int64      `gorm:"not null" json:"size_bytes"`
	ExpiresAt       *time.Time `gorm:"index" json:"expires_at"`
	CreatedAt       time.Time  `gorm:"not null;index" json:"created_at"`
}

func (MediaAsset) TableName() string { return "media_assets" }
