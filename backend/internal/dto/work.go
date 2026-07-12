package dto

// PublishWorkRequest 发布作品请求
type PublishWorkRequest struct {
	TaskID   int64  `json:"task_id" binding:"required"`
	Title    string `json:"title" binding:"required,max=255"`
	Category string `json:"category" binding:"omitempty,max=50"`
}

// WorkResponse 作品响应
type WorkResponse struct {
	ID           int64  `json:"id"`
	UserID       int64  `json:"user_id"`
	User         *UserInfo `json:"user,omitempty"`
	Title        string `json:"title"`
	Prompt       string `json:"prompt"`
	Category     string `json:"category"`
	Type         string `json:"type"`
	Ratio        string `json:"ratio"`
	ImageURL     string `json:"image_url"`
	VideoURL     string `json:"video_url"`
	AuditStatus  string `json:"audit_status"`
	LikesCount   int    `json:"likes_count"`
	CollectsCount int   `json:"collects_count"`
	ViewsCount   int    `json:"views_count"`
	IsLiked      bool   `json:"is_liked"`
	IsCollected  bool   `json:"is_collected"`
	PublishedAt  string `json:"published_at"`
	CreatedAt    string `json:"created_at"`
}

// WorkListRequest 作品列表请求
type WorkListRequest struct {
	Category string `form:"category"`
	Type     string `form:"type"`
	UserID   int64  `form:"user_id"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// WorkAuditRequest 作品审核请求
type WorkAuditRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
	Reason string `json:"reason" binding:"omitempty,max=500"`
}
