package model

import "time"

type User struct {
	BaseModel
	Email        *string    `gorm:"column:email" json:"email,omitempty"`
	Phone        *string    `gorm:"column:phone" json:"phone,omitempty"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	Nickname     string     `gorm:"column:nickname" json:"nickname"`
	AvatarURL    string     `gorm:"column:avatar_url" json:"avatar_url"`
	Credits      int64      `gorm:"column:credits" json:"credits"`
	Status       string     `gorm:"column:status" json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
}

func (User) TableName() string {
	return "users"
}

type APIKey struct {
	BaseModel
	UserID     uint64     `gorm:"column:user_id" json:"user_id"`
	Name       string     `gorm:"column:name" json:"name"`
	KeyPrefix  string     `gorm:"column:key_prefix" json:"key_prefix"`
	KeyHash    string     `gorm:"column:key_hash" json:"-"`
	Status     string     `gorm:"column:status" json:"status"`
	LastUsedAt *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
