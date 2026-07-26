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
	Phone        *string        `gorm:"size:20;uniqueIndex" json:"phone"`
	CreatedAt    time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserSession is a server-side record for a signed-in device. JWTs reference
// this record so a user can explicitly revoke a device before token expiry.
type UserSession struct {
	ID        string     `gorm:"size:64;primaryKey" json:"id"`
	UserID    int64      `gorm:"not null;index" json:"user_id"`
	Device    string     `gorm:"size:255;not null" json:"device"`
	IP        string     `gorm:"size:64" json:"ip"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time  `gorm:"not null" json:"updated_at"`
}

func (UserSession) TableName() string { return "user_sessions" }

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
