package model

import "time"

// StorageConfig 存储配置
type StorageConfig struct {
	ID        int64     `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`        // 配置名称
	Type      string    `gorm:"size:50;not null" json:"type"`         // 存储类型: local, tencent_cos, aliyun_oss, cloudflare
	LocalPath string    `gorm:"size:255" json:"local_path"`           // 本地存储路径
	Endpoint  string    `gorm:"size:255" json:"endpoint"`             // 端点
	AccessKey string    `gorm:"size:255" json:"access_key"`           // AccessKey
	SecretKey string    `gorm:"size:255" json:"secret_key"`           // SecretKey
	Bucket    string    `gorm:"size:100" json:"bucket"`               // 桶名称
	Region    string    `gorm:"size:50" json:"region"`                // 区域
	Domain    string    `gorm:"size:255" json:"domain"`               // CDN 域名
	IsEnabled bool      `gorm:"default:false;not null" json:"is_enabled"` // 是否启用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (StorageConfig) TableName() string {
	return "storage_configs"
}
