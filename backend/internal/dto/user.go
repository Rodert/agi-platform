package dto

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Name   string `json:"name" binding:"omitempty,max=50"`
	Avatar string `json:"avatar" binding:"omitempty,url"`
	Bio    string `json:"bio" binding:"omitempty,max=500"`
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
}
