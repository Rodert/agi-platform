package dto

// StorageConfigRequest 存储配置请求
type StorageConfigRequest struct {
	Name      string `json:"name" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=local tencent_cos aliyun_oss cloudflare"`
	LocalPath string `json:"local_path"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Domain    string `json:"domain"`
}

// StorageConfigResponse 存储配置响应
type StorageConfigResponse struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	LocalPath string `json:"local_path"`
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"` // 前端显示时需要脱敏
	Bucket    string `json:"bucket"`
	Region    string `json:"region"`
	Domain    string `json:"domain"`
	IsEnabled bool   `json:"is_enabled"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
