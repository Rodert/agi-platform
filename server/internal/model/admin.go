package model

import "time"

type AdminUser struct {
	BaseModel
	Username     string     `gorm:"column:username" json:"username"`
	PasswordHash string     `gorm:"column:password_hash" json:"-"`
	Nickname     string     `gorm:"column:nickname" json:"nickname"`
	Role         string     `gorm:"column:role" json:"role"`
	Status       string     `gorm:"column:status" json:"status"`
	LastLoginAt  *time.Time `gorm:"column:last_login_at" json:"last_login_at,omitempty"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

type AdminOperationLog struct {
	ID         uint64    `gorm:"primaryKey;column:id" json:"id"`
	AdminID    *uint64   `gorm:"column:admin_id" json:"admin_id,omitempty"`
	Action     string    `gorm:"column:action" json:"action"`
	TargetType string    `gorm:"column:target_type" json:"target_type"`
	TargetID   *uint64   `gorm:"column:target_id" json:"target_id,omitempty"`
	Detail     []byte    `gorm:"column:detail" json:"detail,omitempty"`
	IP         string    `gorm:"column:ip" json:"ip"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (AdminOperationLog) TableName() string {
	return "admin_operation_logs"
}
