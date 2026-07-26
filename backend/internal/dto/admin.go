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

type AdminProfileResponse struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	LastLoginAt string `json:"last_login_at"`
	LastLoginIP string `json:"last_login_ip"`
	CreatedAt   string `json:"created_at"`
}

type AdminUpdateProfileRequest struct {
	Name            string `json:"name" binding:"required,min=1,max=50"`
	CurrentPassword string `json:"current_password" binding:"omitempty,min=6,max=128"`
	NewPassword     string `json:"new_password" binding:"omitempty,min=8,max=128"`
}

type AdminManagerResponse struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	IsActive    bool   `json:"is_active"`
	LastLoginAt string `json:"last_login_at"`
	CreatedAt   string `json:"created_at"`
}

type CreateAdminManagerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Name     string `json:"name" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Role     string `json:"role" binding:"required,oneof=admin auditor"`
}

type UpdateAdminManagerRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=50"`
	Password string `json:"password" binding:"omitempty,min=8,max=128"`
	Role     string `json:"role" binding:"required,oneof=admin auditor"`
	IsActive *bool  `json:"is_active"`
}

// AdminWorkAuditRequest 作品审核请求
type AdminWorkAuditRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Reason string `json:"reason"`
}

// AdminWorkListRequest 管理端作品列表筛选条件。
type AdminWorkListRequest struct {
	Status   string `form:"status" binding:"omitempty,oneof=pending approved rejected offline"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AdminWorkStatusRequest 作品公开状态变更。下架不会删除已提升的长期资源。
type AdminWorkStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=approved offline"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}

// AdminLogListRequest describes filters for administrative audit logs.
// Dates use the admin UI format: 2006-01-02 15:04:05.
type AdminLogListRequest struct {
	Operator string `form:"operator" binding:"omitempty,max=50"`
	Action   string `form:"action" binding:"omitempty,max=50"`
	StartAt  string `form:"start_at" binding:"omitempty,max=32"`
	EndAt    string `form:"end_at" binding:"omitempty,max=32"`
	LoginOnly bool   `form:"login_only"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// AdminLogResponse 管理员日志响应
type AdminLogResponse struct {
	ID          int64      `json:"id"`
	AdminID     int64      `json:"admin_id"`
	Action      string     `json:"action"`
	TargetType  string     `json:"target_type"`
	TargetID    int64      `json:"target_id"`
	BeforeData  string     `json:"before_data"`
	AfterData   string     `json:"after_data"`
	Description string     `json:"description"`
	IP          string     `json:"ip"`
	CreatedAt   string     `json:"created_at"`
	Admin       *AdminInfo `json:"admin,omitempty"`
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

// DatabaseTableResponse describes a table discovered from the active schema.
// It is intentionally metadata-only; rows are loaded through a separate paged endpoint.
type DatabaseTableResponse struct {
	Name    string `json:"name"`
	Comment string `json:"comment"`
}

type DatabaseColumnResponse struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primary_key"`
}

type DatabaseTableDataResponse struct {
	Table   string                   `json:"table"`
	Columns []DatabaseColumnResponse `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int64                    `json:"total"`
	Page    int                      `json:"page"`
	PageSize int                     `json:"page_size"`
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

// AdminReportRequest defines the inclusive date range for operational reports.
type AdminReportRequest struct {
	StartDate string `form:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate   string `form:"end_date" binding:"required,datetime=2006-01-02"`
}

type AdminReportSummary struct {
	NewUsers        int64   `json:"new_users"`
	ActiveUsers     int64   `json:"active_users"`
	Tasks           int64   `json:"tasks"`
	SuccessTasks    int64   `json:"success_tasks"`
	FailedTasks     int64   `json:"failed_tasks"`
	PendingTasks    int64   `json:"pending_tasks"`
	SuccessRate     float64 `json:"success_rate"`
	CreditsConsumed int64   `json:"credits_consumed"`
	Works           int64   `json:"works"`
	PendingWorks    int64   `json:"pending_works"`
	ApprovedWorks   int64   `json:"approved_works"`
	RejectedWorks   int64   `json:"rejected_works"`
	OfflineWorks    int64   `json:"offline_works"`
}

type AdminReportDailyPoint struct {
	Date            string `json:"date"`
	NewUsers        int64  `json:"new_users"`
	ActiveUsers     int64  `json:"active_users"`
	Tasks           int64  `json:"tasks"`
	SuccessTasks    int64  `json:"success_tasks"`
	FailedTasks     int64  `json:"failed_tasks"`
	CreditsConsumed int64  `json:"credits_consumed"`
	Works           int64  `json:"works"`
}

type AdminReportBreakdownItem struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
}

type AdminReportResponse struct {
	StartDate    string                     `json:"start_date"`
	EndDate      string                     `json:"end_date"`
	Summary      AdminReportSummary         `json:"summary"`
	Daily        []AdminReportDailyPoint    `json:"daily"`
	TaskTypes    []AdminReportBreakdownItem `json:"task_types"`
	TaskModels   []AdminReportBreakdownItem `json:"task_models"`
	TaskChannels []AdminReportBreakdownItem `json:"task_channels"`
	WorkStatuses []AdminReportBreakdownItem `json:"work_statuses"`
	WorkCategories []AdminReportBreakdownItem `json:"work_categories"`
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

type AdminCreditLedgerResponse struct {
	ID           int64  `json:"id"`
	Type         string `json:"type"`
	Amount       int    `json:"amount"`
	Title        string `json:"title"`
	SourceType   string `json:"source_type"`
	SourceID     int64  `json:"source_id"`
	BalanceAfter int    `json:"balance_after"`
	CreatedAt    string `json:"created_at"`
}

type AdminCreditLedgerListRequest struct {
	Type       string `form:"type" binding:"omitempty,oneof=income expense"`
	SourceType string `form:"source_type" binding:"omitempty,max=50"`
	StartAt    string `form:"start_at" binding:"omitempty,max=32"`
	EndAt      string `form:"end_at" binding:"omitempty,max=32"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AdminCreateRedeemCodesRequest struct {
	BatchName string `json:"batch_name" binding:"omitempty,max=100"`
	Amount    int    `json:"amount" binding:"required,min=1,max=1000000"`
	Quantity  int    `json:"quantity" binding:"required,min=1,max=1000"`
	ExpiresAt string `json:"expires_at" binding:"omitempty,datetime=2006-01-02 15:04:05"`
}

type AdminRedeemCodeListRequest struct {
	Keyword  string `form:"keyword" binding:"omitempty,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=unused used expired"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AdminRedeemCodeResponse struct {
	ID        int64  `json:"id"`
	Code      string `json:"code"`
	Amount    int    `json:"amount"`
	BatchID   string `json:"batch_id"`
	BatchName string `json:"batch_name"`
	UsedBy    int64  `json:"used_by"`
	UsedName  string `json:"used_name"`
	UsedEmail string `json:"used_email"`
	UsedAt    string `json:"used_at"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
}

type AdminCreditPackageRequest struct {
	ID          int64   `json:"id" binding:"required"`
	Name        string  `json:"name" binding:"required,min=1,max=50"`
	Price       float64 `json:"price" binding:"required,gt=0,max=1000000"`
	Points      int     `json:"points" binding:"required,min=1,max=1000000"`
	Note        string  `json:"note" binding:"omitempty,max=255"`
	PurchaseURL string  `json:"purchase_url" binding:"omitempty,url,max=500"`
	IsHot       bool    `json:"is_hot"`
	IsActive    bool    `json:"is_active"`
}
