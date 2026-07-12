package dto

// CreateImageTaskRequest 创建图片生成任务请求
type CreateImageTaskRequest struct {
	Prompt         string                 `json:"prompt" binding:"required,min=1,max=2000"`
	ModelName      string                 `json:"model_name" binding:"required"`
	Params         map[string]interface{} `json:"params"`
	ReferenceImage string                 `json:"reference_image"` // Base64 或 URL
}

// CreateVideoTaskRequest 创建视频生成任务请求
type CreateVideoTaskRequest struct {
	Prompt         string                 `json:"prompt" binding:"required,min=1,max=2000"`
	ModelName      string                 `json:"model_name" binding:"required"`
	Params         map[string]interface{} `json:"params"`
	FirstFrameURL  string                 `json:"first_frame_url"`
	LastFrameURL   string                 `json:"last_frame_url"`
}

// CreateProductTaskRequest 创建商品图生成任务请求
type CreateProductTaskRequest struct {
	Prompt            string                 `json:"prompt" binding:"required,min=1,max=2000"`
	ModelName         string                 `json:"model_name" binding:"required"`
	Params            map[string]interface{} `json:"params"`
	ProductImageURL   string                 `json:"product_image_url" binding:"required"`
}

// TaskResponse 任务响应
type TaskResponse struct {
	ID           int64                  `json:"id"`
	Title        string                 `json:"title"`
	Type         string                 `json:"type"`
	Status       string                 `json:"status"`
	Progress     int                    `json:"progress"`
	Prompt       string                 `json:"prompt"`
	ModelName    string                 `json:"model_name"`
	ResultURL    string                 `json:"result_url"`
	ThumbnailURL string                 `json:"thumbnail_url"`
	ErrorMsg     string                 `json:"error_msg"`
	Cost         int                    `json:"cost"`
	CreatedAt    string                 `json:"created_at"`
	CompletedAt  string                 `json:"completed_at"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	Status   string `form:"status"`                             // queued/processing/success/failed
	Type     string `form:"type"`                               // image/video/product
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

// GetModelsResponse 获取模型列表响应
type GetModelsResponse struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Type         string                 `json:"type"`
	Provider     string                 `json:"provider"`
	Description  string                 `json:"description"`
	LogoURL      string                 `json:"logo_url"`
	Tag          string                 `json:"tag"`
	Cost         int                    `json:"cost"`
	ParamsConfig map[string]interface{} `json:"params_config"`
}
