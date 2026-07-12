package model

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser 管理员
type AdminUser struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username      string         `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"`
	Name          string         `gorm:"size:50;not null" json:"name"`
	Role          string         `gorm:"size:20;not null" json:"role"` // super_admin/admin/auditor
	Permissions   string         `gorm:"type:json" json:"permissions"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	LastLoginAt   *time.Time     `json:"last_login_at"`
	LastLoginIP   string         `gorm:"size:50" json:"last_login_ip"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

// AdminLog 管理员操作日志
type AdminLog struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	AdminID    int64     `gorm:"not null;index" json:"admin_id"`
	Action     string    `gorm:"size:50;not null;index" json:"action"`
	TargetType string    `gorm:"size:50" json:"target_type"`
	TargetID   int64     `json:"target_id"`
	BeforeData string    `gorm:"type:json;not null;default:'{}'" json:"before_data"`
	AfterData  string    `gorm:"type:json;not null;default:'{}'" json:"after_data"`
	Description string   `gorm:"size:500" json:"description"`
	IP         string    `gorm:"size:50" json:"ip"`
	CreatedAt  time.Time `gorm:"not null;index" json:"created_at"`
}

func (AdminLog) TableName() string {
	return "admin_logs"
}
