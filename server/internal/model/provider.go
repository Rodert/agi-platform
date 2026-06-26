package model

import "time"

type Provider struct {
	BaseModel
	Code           string `gorm:"column:code" json:"code"`
	Name           string `gorm:"column:name" json:"name"`
	Type           string `gorm:"column:type" json:"type"`
	BaseURL        string `gorm:"column:base_url" json:"base_url"`
	Enabled        bool   `gorm:"column:enabled" json:"enabled"`
	TimeoutSeconds int    `gorm:"column:timeout_seconds" json:"timeout_seconds"`
	RetryCount     int    `gorm:"column:retry_count" json:"retry_count"`
	Priority       int    `gorm:"column:priority" json:"priority"`
	DailyLimit     *int64 `gorm:"column:daily_limit" json:"daily_limit,omitempty"`
	Remark         string `gorm:"column:remark" json:"remark"`
}

func (Provider) TableName() string {
	return "providers"
}

type ProviderKey struct {
	BaseModel
	ProviderID      uint64     `gorm:"column:provider_id" json:"provider_id"`
	Name            string     `gorm:"column:name" json:"name"`
	APIKeyEncrypted string     `gorm:"column:api_key_encrypted" json:"-"`
	Status          string     `gorm:"column:status" json:"status"`
	Weight          int        `gorm:"column:weight" json:"weight"`
	DailyLimit      *int64     `gorm:"column:daily_limit" json:"daily_limit,omitempty"`
	DailyUsed       int64      `gorm:"column:daily_used" json:"daily_used"`
	LastUsedAt      *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
	LastError       string     `gorm:"column:last_error" json:"last_error"`
}

func (ProviderKey) TableName() string {
	return "provider_keys"
}
