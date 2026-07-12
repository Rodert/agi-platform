package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID           int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Email        string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	Name         string         `gorm:"size:50;not null" json:"name"`
	Avatar       string         `gorm:"size:255" json:"avatar"`
	Bio          string         `gorm:"size:500" json:"bio"`
	Level        string         `gorm:"size:20;default:free" json:"level"` // free/member/pro
	InviteCode   string         `gorm:"size:10;uniqueIndex;not null" json:"invite_code"`
	InvitedBy    int64          `gorm:"index" json:"invited_by"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// VerificationCode 验证码模型
type VerificationCode struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"size:100;index;not null" json:"email"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	Type      string    `gorm:"size:20;not null" json:"type"` // register/login/reset
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (VerificationCode) TableName() string {
	return "verification_codes"
}
