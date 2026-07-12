package dto

import "github.com/javapub/agi-platform-backend/internal/model"

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token string    `json:"token"`
	Admin *AdminInfo `json:"admin"`
}

// AdminInfo 管理员信息
type AdminInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// AdminWorkAuditRequest 作品审核请求
type AdminWorkAuditRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Reason string `json:"reason"`
}

// AdminUserListRequest 用户列表请求
type AdminUserListRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AdminTaskListRequest 任务列表请求
type AdminTaskListRequest struct {
	Status   string `form:"status"`
	Type     string `form:"type"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AdminStatsResponse 统计响应
type AdminStatsResponse struct {
	TotalUsers      int64 `json:"total_users"`
	TotalTasks      int64 `json:"total_tasks"`
	TotalWorks      int64 `json:"total_works"`
	PendingWorks    int64 `json:"pending_works"`
	TodayUsers      int64 `json:"today_users"`
	TodayTasks      int64 `json:"today_tasks"`
	TodayWorks      int64 `json:"today_works"`
}

type UserListRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type UserListResponse struct {
	List  []*model.User `json:"list"`
	Total int64         `json:"total"`
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type AdminUpdateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"omitempty,min=8"`
	Level    string `json:"level" binding:"required,oneof=free member pro"`
}

type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active"`
}
