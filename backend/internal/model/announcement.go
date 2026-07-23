package model

import "time"

// Announcement is an admin-authored platform announcement visible to every user.
// It intentionally has no recipient or read-state relationship.
type Announcement struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"size:120;not null" json:"title"`
	Content     string     `gorm:"type:text;not null" json:"content"`
	Category    string     `gorm:"size:30;not null;default:system" json:"category"`
	IsPublished bool       `gorm:"not null;default:false;index" json:"is_published"`
	PublishedAt *time.Time `gorm:"index" json:"published_at"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null" json:"updated_at"`
}

func (Announcement) TableName() string { return "announcements" }
