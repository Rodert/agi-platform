package model

import (
	"time"
)

// CreditAccount 灵感值账户
type CreditAccount struct {
	ID           int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       int64     `gorm:"not null;uniqueIndex" json:"user_id"`
	Balance      int       `gorm:"not null;default:0" json:"balance"`
	TotalIncome  int       `gorm:"not null;default:0" json:"total_income"`
	TotalExpense int       `gorm:"not null;default:0" json:"total_expense"`
	CreatedAt    time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt    time.Time `gorm:"not null" json:"updated_at"`
}

func (CreditAccount) TableName() string {
	return "credit_accounts"
}

// CreditLedger 灵感值流水
type CreditLedger struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         int64     `gorm:"not null;index" json:"user_id"`
	Type           string    `gorm:"size:20;not null" json:"type"` // income/expense
	Amount         int       `gorm:"not null" json:"amount"`
	Title          string    `gorm:"size:255;not null" json:"title"`
	SourceType     string    `gorm:"size:50" json:"source_type"` // recharge/task/checkin/redeem/gift/invite_register/invite_recharge
	SourceID       int64     `json:"source_id"`
	BalanceAfter   int       `gorm:"not null" json:"balance_after"`
	IdempotencyKey string    `gorm:"size:100;uniqueIndex" json:"idempotency_key"`
	CreatedAt      time.Time `gorm:"not null;index" json:"created_at"`
}

func (CreditLedger) TableName() string {
	return "credit_ledgers"
}

// CheckinRecord 签到记录
type CheckinRecord struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"not null;uniqueIndex:idx_user_date" json:"user_id"`
	Date      string    `gorm:"size:10;not null;uniqueIndex:idx_user_date" json:"date"` // YYYY-MM-DD
	Streak    int       `gorm:"not null;default:1" json:"streak"`
	Reward    int       `gorm:"not null" json:"reward"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (CheckinRecord) TableName() string {
	return "checkin_records"
}

// RedeemCode 兑换码
type RedeemCode struct {
	ID        int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Code      string     `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Amount    int        `gorm:"not null" json:"amount"`
	BatchID   string     `gorm:"size:50;index" json:"batch_id"`
	BatchName string     `gorm:"size:100" json:"batch_name"`
	UsedBy    int64      `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `gorm:"not null" json:"created_at"`
}

func (RedeemCode) TableName() string {
	return "redeem_codes"
}

// CreditPackage 充值套餐
type CreditPackage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Price     float64   `gorm:"not null" json:"price"`
	Points    int       `gorm:"not null" json:"points"`
	Note      string    `gorm:"size:255" json:"note"`
	IsHot     bool      `gorm:"default:false" json:"is_hot"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

func (CreditPackage) TableName() string {
	return "credit_packages"
}
