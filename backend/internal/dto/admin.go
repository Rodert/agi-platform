package dto

import "time"

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AdminLoginResponse 管理员登录响应
type AdminLoginResponse struct {
	Token string     `json:"token"`
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
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Type     string `form:"type"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AdminTaskResponse struct {
	ID               int64                      `json:"id"`
	UserID           int64                      `json:"user_id"`
	UserName         string                     `json:"user_name"`
	UserEmail        string                     `json:"user_email"`
	ChannelName      string                     `json:"channel_name"`
	ProviderTaskID   string                     `json:"provider_task_id"`
	ProviderStatus   string                     `json:"provider_status"`
	ProviderResponse map[string]interface{}     `json:"provider_response"`
	LastPolledAt     string                     `json:"last_polled_at"`
	ModelName        string                     `json:"model_name"`
	Type             string                     `json:"type"`
	Status           string                     `json:"status"`
	Progress         int                        `json:"progress"`
	Prompt           string                     `json:"prompt"`
	Params           map[string]interface{}     `json:"params"`
	Cost             int                        `json:"cost"`
	ResultURL        string                     `json:"result_url"`
	ThumbnailURL     string                     `json:"thumbnail_url"`
	ErrorMsg         string                     `json:"error_msg"`
	AttemptCount     int                        `json:"attempt_count"`
	MaxRetryAttempts int                        `json:"max_retry_attempts"`
	LastRetryAt      string                     `json:"last_retry_at"`
	Attempts         []AdminTaskAttemptResponse `json:"attempts"`
	Assets           []AdminMediaAssetResponse  `json:"assets"`
	CreatedAt        string                     `json:"created_at"`
	CompletedAt      string                     `json:"completed_at"`
}

type AdminMediaAssetResponse struct {
	ResourceType    string `json:"resource_type"`
	StorageConfigID int64  `json:"storage_config_id"`
	ObjectKey       string `json:"object_key"`
	PublicURL       string `json:"public_url"`
	ContentType     string `json:"content_type"`
	SizeBytes       int64  `json:"size_bytes"`
	ExpiresAt       string `json:"expires_at"`
}

type AdminTaskAttemptResponse struct {
	Attempt     int    `json:"attempt"`
	Status      string `json:"status"`
	ErrorMsg    string `json:"error_msg"`
	StartedAt   string `json:"started_at"`
	CompletedAt string `json:"completed_at"`
}

// AdminStatsResponse 统计响应
type AdminStatsResponse struct {
	TotalUsers   int64 `json:"total_users"`
	TotalTasks   int64 `json:"total_tasks"`
	TotalWorks   int64 `json:"total_works"`
	PendingWorks int64 `json:"pending_works"`
	TodayUsers   int64 `json:"today_users"`
	TodayTasks   int64 `json:"today_tasks"`
	TodayWorks   int64 `json:"today_works"`
}

type UserListRequest struct {
	Username string `form:"username"`
	Email    string `form:"email"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type UserListResponse struct {
	List  []*AdminUserResponse `json:"list"`
	Total int64                `json:"total"`
}

type AdminUserResponse struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Balance   int       `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
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

type AdminRechargeCreditRequest struct {
	Type   string `json:"type" binding:"required,oneof=add deduct"`
	Amount int    `json:"amount" binding:"required,min=1,max=1000000"`
	Remark string `json:"remark" binding:"required,min=1,max=200"`
}

type AdminRechargeCreditResponse struct {
	Balance int `json:"balance"`
}
