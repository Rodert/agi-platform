package model

import "time"

const (
	WalletLogTypeConsume = "consume"
	WalletLogTypeRefund  = "refund"
)

type WalletLog struct {
	ID            uint64    `gorm:"primaryKey;column:id" json:"id"`
	UserID        uint64    `gorm:"column:user_id" json:"user_id"`
	Type          string    `gorm:"column:type" json:"type"`
	Amount        int64     `gorm:"column:amount" json:"amount"`
	BalanceBefore int64     `gorm:"column:balance_before" json:"balance_before"`
	BalanceAfter  int64     `gorm:"column:balance_after" json:"balance_after"`
	RelatedType   string    `gorm:"column:related_type" json:"related_type"`
	RelatedID     *uint64   `gorm:"column:related_id" json:"related_id,omitempty"`
	Remark        string    `gorm:"column:remark" json:"remark"`
	OperatorType  string    `gorm:"column:operator_type" json:"operator_type"`
	OperatorID    *uint64   `gorm:"column:operator_id" json:"operator_id,omitempty"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (WalletLog) TableName() string {
	return "wallet_logs"
}
