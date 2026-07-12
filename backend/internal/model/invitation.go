package model

import (
	"time"
)

// Invitation 邀请记录
type Invitation struct {
	ID           int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	InviterID    int64      `gorm:"not null;index" json:"inviter_id"`
	InviteeID    int64      `gorm:"not null;index" json:"invitee_id"`
	InviteCode   string     `gorm:"size:10;index;not null" json:"invite_code"`
	Status       string     `gorm:"size:20;not null" json:"status"` // registered/rewarded
	RegisteredAt time.Time  `gorm:"not null" json:"registered_at"`
	RewardedAt   *time.Time `json:"rewarded_at"`
	CreatedAt    time.Time  `gorm:"not null" json:"created_at"`
}

func (Invitation) TableName() string {
	return "invitations"
}

// InvitationReward 邀请奖励记录
type InvitationReward struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	InvitationID  int64     `gorm:"not null;index" json:"invitation_id"`
	InviterID     int64     `gorm:"not null;index" json:"inviter_id"`
	InviteeID     int64     `gorm:"not null;index" json:"invitee_id"`
	InviterReward int       `gorm:"not null" json:"inviter_reward"`
	InviteeReward int       `gorm:"not null" json:"invitee_reward"`
	TriggerType   string    `gorm:"size:20;not null" json:"trigger_type"` // register/first_recharge
	CreatedAt     time.Time `gorm:"not null" json:"created_at"`
}

func (InvitationReward) TableName() string {
	return "invitation_rewards"
}
