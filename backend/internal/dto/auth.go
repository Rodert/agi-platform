package dto

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Code           string `json:"code" binding:"omitempty,len=6"`
	Password       string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
	InviteCode     string `json:"invite_code"` // 可选
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password,omitempty"`
	Code     string `json:"code,omitempty"`
	Type     string `json:"type" binding:"required,oneof=password code"` // password/code
}

// SendCodeRequest 发送验证码请求
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register login reset"` // register/login/reset
}

// AuthResponse 认证响应
type AuthResponse struct {
	Token string    `json:"token"`
	User  *UserInfo `json:"user"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID         int64  `json:"id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Level      string `json:"level"`
	InviteCode string `json:"invite_code"`
}
