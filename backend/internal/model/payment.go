package model

import (
	"time"

	"gorm.io/datatypes"
)

// PaymentOrder 支付订单
type PaymentOrder struct {
	ID            int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int64      `gorm:"not null;index" json:"user_id"`
	OrderNo       string     `gorm:"size:50;uniqueIndex;not null" json:"order_no"`
	PackageID     int64      `gorm:"not null" json:"package_id"`
	Amount        float64    `gorm:"not null" json:"amount"`
	Points        int        `gorm:"not null" json:"points"`
	Status        string     `gorm:"size:20;not null;default:pending" json:"status"` // pending/paid/failed/refunded
	ChannelID     int64      `gorm:"not null;index" json:"channel_id"`
	PayMethodName string     `gorm:"size:50" json:"pay_method_name"`
	TradeNo       string     `gorm:"size:100" json:"trade_no"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt     time.Time  `gorm:"not null" json:"updated_at"`
}

func (PaymentOrder) TableName() string {
	return "payment_orders"
}

// PaymentTransaction 支付交易记录
type PaymentTransaction struct {
	ID            int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID       int64          `gorm:"not null;index" json:"order_id"`
	ChannelID     int64          `gorm:"not null" json:"channel_id"`
	TransactionID string         `gorm:"size:100;index" json:"transaction_id"`
	Amount        float64        `gorm:"not null" json:"amount"`
	Status        string         `gorm:"size:20;not null" json:"status"`
	RequestData   datatypes.JSON `gorm:"type:json" json:"request_data"`
	CallbackData  datatypes.JSON `gorm:"type:json" json:"callback_data"`
	CreatedAt     time.Time      `gorm:"not null" json:"created_at"`
}

func (PaymentTransaction) TableName() string {
	return "payment_transactions"
}

// PaymentChannel 支付渠道配置
type PaymentChannel struct {
	ID          int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"size:50;not null" json:"name"`
	ChannelType string         `gorm:"size:20;not null" json:"channel_type"` // alipay/wechat/epay/demo
	MerchantID  string         `gorm:"size:100" json:"merchant_id"`
	Config      datatypes.JSON `gorm:"type:json;not null" json:"config"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
}

func (PaymentChannel) TableName() string {
	return "payment_channels"
}
