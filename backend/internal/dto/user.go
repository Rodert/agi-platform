package dto

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,max=50"`
	Avatar string `json:"avatar" binding:"omitempty,url"`
	Bio    string `json:"bio" binding:"omitempty,max=500"`
}

type BindPhoneRequest struct {
	Phone string `json:"phone" binding:"required,e164"`
	Code string `json:"code" binding:"required,len=6"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8,max=128"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=128"`
}

type UserSessionResponse struct {
	ID string `json:"id"`
	Device string `json:"device"`
	IP string `json:"ip"`
	CreatedAt string `json:"created_at"`
	Current bool `json:"current"`
}

// UserProfileResponse 用户资料响应
type UserProfileResponse struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Bio        string `json:"bio"`
	Level      string `json:"level"`
	Balance    int    `json:"balance"`
	InviteCode string `json:"invite_code"`
	Following  int    `json:"following"`
	Followers  int    `json:"followers"`
	CreatedAt  string `json:"created_at"`
	Phone      string `json:"phone"`
}

// CreditLedgerListRequest is the current user's credit-history pagination.
type CreditLedgerListRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// CreditLedgerResponse is a user-safe representation of one credit record.
type CreditLedgerResponse struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"`
	Amount       int    `json:"amount"`
	Title        string `json:"title"`
	SourceType   string `json:"source_type"`
	BalanceAfter int    `json:"balance_after"`
	CreatedAt    string `json:"created_at"`
}

type RedeemCodeRequest struct {
	Code string `json:"code" binding:"required,min=4,max=50"`
}

type RedeemCodeResponse struct {
	Amount  int `json:"amount"`
	Balance int `json:"balance"`
}

type CreditPackageResponse struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Points      int     `json:"points"`
	Note        string  `json:"note"`
	PurchaseURL string  `json:"purchase_url"`
	IsHot       bool    `json:"is_hot"`
}
