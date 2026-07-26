package model

import (
	"time"
)

// Work 作品模型
type Work struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64      `gorm:"not null;index" json:"user_id"`
	TaskID        int64      `gorm:"uniqueIndex" json:"task_id"`
	Title         string     `gorm:"size:255;not null" json:"title"`
	Prompt        string     `gorm:"type:text;not null" json:"prompt"`
	Category      string     `gorm:"size:50;index" json:"category"`
	Type          string     `gorm:"size:20;not null" json:"type"` // image/video
	Ratio         string     `gorm:"size:10" json:"ratio"`
	ImageURL      string     `gorm:"size:500" json:"image_url"`
	VideoURL      string     `gorm:"size:500" json:"video_url"`
	AuditStatus   string     `gorm:"size:20;not null;default:pending;index" json:"audit_status"` // pending/approved/rejected/offline
	AuditReason   string     `gorm:"size:500" json:"audit_reason"`
	AuditAdminID  int64      `json:"audit_admin_id"`
	AuditedAt     *time.Time `json:"audited_at"`
	LikesCount    int        `gorm:"default:0" json:"likes_count"`
	CollectsCount int        `gorm:"default:0" json:"collects_count"`
	ViewsCount    int        `gorm:"default:0" json:"views_count"`
	PublishedAt   *time.Time `json:"published_at"`
	CreatedAt     time.Time  `gorm:"not null;index" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null" json:"updated_at"`
}

func (Work) TableName() string {
	return "works"
}

// WorkAudit 作品审核记录
type WorkAudit struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkID     int64     `gorm:"not null;index" json:"work_id"`
	AdminID    int64     `gorm:"not null;index" json:"admin_id"`
	Status     string    `gorm:"size:20;not null" json:"status"` // approved/rejected/offline
	Reason     string    `gorm:"size:500" json:"reason"`
	AuditedAt  time.Time `gorm:"not null" json:"audited_at"`
	CreatedAt  time.Time `gorm:"not null" json:"created_at"`
}

func (WorkAudit) TableName() string {
	return "work_audits"
}

// WorkLike 作品点赞
type WorkLike struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_work" json:"user_id"`
	WorkID    int64     `gorm:"not null;uniqueIndex:idx_user_work;index" json:"work_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (WorkLike) TableName() string {
	return "work_likes"
}

// WorkCollect 作品收藏
type WorkCollect struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_work" json:"user_id"`
	WorkID    int64     `gorm:"not null;uniqueIndex:idx_user_work;index" json:"work_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (WorkCollect) TableName() string {
	return "work_collects"
}
