package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"github.com/javapub/agi-platform-backend/pkg/response"
	"go.uber.org/zap"
)

type AdminConfigHandler struct {
	configRepo            *repository.ConfigRepository
	modelRepo             *repository.AIModelRepository
	providerRepo          *repository.AIProviderAccountRepository
	channelModelRepo      *repository.ChannelModelRepository
	channelCatalogService *service.ChannelCatalogService
	storageService        *service.StorageConfigService
	resourcePolicyService *service.ResourcePolicyService
	emailService          *service.EmailService
	adminRepo             *repository.AdminRepository
	redeemCodeService     *service.RedeemCodeService
}

func NewAdminConfigHandler(c *repository.ConfigRepository, m *repository.AIModelRepository, p *repository.AIProviderAccountRepository, cm *repository.ChannelModelRepository, catalog *service.ChannelCatalogService, s *service.StorageConfigService, rp *service.ResourcePolicyService, e *service.EmailService, a *repository.AdminRepository, redeem *service.RedeemCodeService) *AdminConfigHandler {
	return &AdminConfigHandler{configRepo: c, modelRepo: m, providerRepo: p, channelModelRepo: cm, channelCatalogService: catalog, storageService: s, resourcePolicyService: rp, emailService: e, adminRepo: a, redeemCodeService: redeem}
}

func (h *AdminConfigHandler) GetCreditPackages(c *gin.Context) {
	items, err := h.redeemCodeService.ListPackages()
	if err != nil { response.Error(c, err); return }
	response.Success(c, items)
}

func (h *AdminConfigHandler) SaveCreditPackages(c *gin.Context) {
	var req []*dto.AdminCreditPackageRequest
	if err := c.ShouldBindJSON(&req); err != nil { response.Error(c, errors.NewWithDetails(errors.ErrCodeBadRequest, "套餐参数错误", err.Error())); return }
	if err := h.redeemCodeService.UpdatePackages(req); err != nil { response.Error(c, err); return }
	h.recordAudit(c, "update_credit_packages", "credit_package", 0, "更新充值套餐", req)
	response.Success(c, nil)
}

func (h *AdminConfigHandler) recordAudit(c *gin.Context, action, targetType string, targetID int64, description string, after interface{}) {
	adminID, ok := c.Get("admin_id")
	if !ok { return }
	afterData, err := json.Marshal(after)
	if err != nil { afterData = []byte("{}") }
	if err := h.adminRepo.CreateLog(&model.AdminLog{AdminID: adminID.(int64), Action: action, TargetType: targetType, TargetID: targetID, AfterData: string(afterData), Description: description, IP: c.ClientIP(), CreatedAt: time.Now()}); err != nil {
		if logger.Log != nil {
			logger.Error("记录管理配置审计日志失败", zap.String("action", action), zap.Error(err))
		}
	}
}

type basicConfigRequest struct {
	SiteName        string `json:"site_name"`
	SiteDesc        string `json:"site_desc"`
	RegisterEnabled bool   `json:"register_enabled"`
}
type userDefaultsRequest struct {
	NewUserGiftAmount       int    `json:"new_user_gift_amount" binding:"min=0,max=1000000"`
	DefaultUserLevel        string `json:"default_user_level" binding:"required,oneof=free member pro"`
	DefaultUserAvatar       string `json:"default_user_avatar" binding:"omitempty,max=255"`
	RegisterEmailVerification bool `json:"register_email_verification"`
}
type emailConfigRequest struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	SMTPSSL      bool   `json:"smtp_ssl"`
	FromName     string `json:"from_name"`
	FromEmail    string `json:"from_email"`
	IsActive     bool   `json:"is_active"`
}
type modelUpdateRequest struct {
	DisplayName  string                 `json:"display_name"`
	Cost         int                    `json:"cost"`
	Description  string                 `json:"description"`
	ParamsConfig map[string]interface{} `json:"params_config"`
}
type statusRequest struct {
	IsActive bool `json:"is_active"`
}
type testEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}
type taskConfigRequest struct {
	MaxActiveTasks   int `json:"max_active_tasks"`
	PromptMaxLength  int `json:"prompt_max_length"`
	MaxRetryAttempts int `json:"max_retry_attempts"`
	ImageConcurrency int `json:"image_concurrency"`
	VideoConcurrency int `json:"video_concurrency"`
}
type promptOptimizationConfigRequest struct {
	IsActive           bool   `json:"is_active"`
	ModelName          string `json:"model_name"`
	SystemPrompt       string `json:"system_prompt"`
	MaxInputLength     int    `json:"max_input_length"`
	CreditCost         int    `json:"credit_cost"`
	RateLimitPerMinute int    `json:"rate_limit_per_minute"`
}
type resourcePolicyRequest struct {
	KeyPrefix     string `json:"key_prefix"`
	RetentionDays int    `json:"retention_days"`
	IsPublic      bool   `json:"is_public"`
	CacheMaxAge   int    `json:"cache_max_age"`
	MaxSizeMB     int    `json:"max_size_mb"`
}

func (h *AdminConfigHandler) GetBasic(c *gin.Context) {
	keys := []string{"site_name", "site_desc", "register_enabled"}
	out := map[string]string{}
	for _, k := range keys {
		if v, e := h.configRepo.GetSystemConfig(k); e == nil {
			out[k] = v.Value
		}
	}
	response.Success(c, out)
}
func (h *AdminConfigHandler) SaveBasic(c *gin.Context) {
	var r basicConfigRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	values := []struct{ k, v, t, d string }{{"site_name", r.SiteName, "string", "网站名称"}, {"site_desc", r.SiteDesc, "string", "网站描述"}, {"register_enabled", strconv.FormatBool(r.RegisterEnabled), "bool", "注册开关"}}
	for _, v := range values {
		if e := h.configRepo.UpsertSystemConfig(v.k, v.v, v.t, "basic", v.d); e != nil {
			response.Error(c, e)
			return
		}
	}
	h.recordAudit(c, "update_basic_config", "system_config", 0, "更新基础配置", r)
	response.Success(c, nil)
}

func (h *AdminConfigHandler) GetUserDefaults(c *gin.Context) {
	amount := 5
	if value, err := h.configRepo.GetSystemConfigValue("new_user_gift_amount", "5"); err == nil {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil && parsed >= 0 { amount = parsed }
	}
	level, _ := h.configRepo.GetSystemConfigValue("default_user_level", "free")
	avatar, _ := h.configRepo.GetSystemConfigValue("default_user_avatar", "")
	verification, _ := h.configRepo.GetSystemConfigValue("register_email_verification", "true")
	response.Success(c, gin.H{"new_user_gift_amount": amount, "default_user_level": level, "default_user_avatar": avatar, "register_email_verification": verification == "true"})
}

func (h *AdminConfigHandler) SaveUserDefaults(c *gin.Context) {
	var req userDefaultsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	values := []struct{ key, value, configType, description string }{
		{"new_user_gift_amount", strconv.Itoa(req.NewUserGiftAmount), "int", "新用户注册礼包灵感值"},
		{"default_user_level", req.DefaultUserLevel, "string", "新用户默认等级"},
		{"default_user_avatar", req.DefaultUserAvatar, "string", "新用户默认头像"},
		{"register_email_verification", strconv.FormatBool(req.RegisterEmailVerification), "bool", "注册时是否需要邮箱验证码"},
	}
	for _, value := range values {
		if err := h.configRepo.UpsertSystemConfig(value.key, value.value, value.configType, "user_defaults", value.description); err != nil {
			response.Error(c, err)
			return
		}
	}
	h.recordAudit(c, "update_user_defaults", "system_config", 0, "更新用户默认设置", req)
	response.Success(c, nil)
}

func (h *AdminConfigHandler) GetPublicRegistrationSettings(c *gin.Context) {
	verification, err := h.configRepo.GetSystemConfigValue("register_email_verification", "true")
	if err != nil { response.Error(c, err); return }
	response.Success(c, gin.H{"register_email_verification": verification == "true"})
}
func (h *AdminConfigHandler) GetEmail(c *gin.Context) {
	v, e := h.configRepo.GetEmailConfig()
	if e != nil {
		response.Error(c, e)
		return
	}
	v.SMTPPassword = ""
	response.Success(c, v)
}
func (h *AdminConfigHandler) SaveEmail(c *gin.Context) {
	var r emailConfigRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	old, e := h.configRepo.GetEmailConfig()
	if e != nil {
		old = &model.EmailConfig{ID: 1}
	}
	if r.SMTPPassword == "" {
		r.SMTPPassword = old.SMTPPassword
	}
	v := &model.EmailConfig{ID: 1, SMTPHost: r.SMTPHost, SMTPPort: r.SMTPPort, SMTPUser: r.SMTPUser, SMTPPassword: r.SMTPPassword, SMTPSSL: r.SMTPSSL, FromName: r.FromName, FromEmail: r.FromEmail, IsActive: r.IsActive}
	if e = h.configRepo.UpdateEmailConfig(v); e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_email_config", "email_config", 1, "更新邮件配置", map[string]interface{}{"smtp_host": v.SMTPHost, "smtp_port": v.SMTPPort, "smtp_user": v.SMTPUser, "smtp_ssl": v.SMTPSSL, "from_name": v.FromName, "from_email": v.FromEmail, "is_active": v.IsActive, "smtp_password_changed": r.SMTPPassword != ""})
	response.Success(c, nil)
}
func (h *AdminConfigHandler) TestEmail(c *gin.Context) {
	var r testEmailRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	if e := h.emailService.SendVerificationCode(r.Email, "123456"); e != nil {
		response.Error(c, e)
		return
	}
	response.Success(c, nil)
}
func (h *AdminConfigHandler) GetTaskConfig(c *gin.Context) {
	v, e := h.configRepo.GetTaskConfig()
	if e != nil {
		response.Error(c, e)
		return
	}
	response.Success(c, v)
}
func (h *AdminConfigHandler) SaveTaskConfig(c *gin.Context) {
	var r taskConfigRequest
	if c.ShouldBindJSON(&r) != nil || r.MaxActiveTasks < 1 || r.MaxActiveTasks > 1000 || r.PromptMaxLength < 1 || r.PromptMaxLength > 50000 || r.MaxRetryAttempts < 0 || r.MaxRetryAttempts > 10 {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "任务配置参数无效"))
		return
	}
	if r.ImageConcurrency == 0 || r.VideoConcurrency == 0 {
		current, err := h.configRepo.GetTaskConfig()
		if err != nil {
			response.Error(c, err)
			return
		}
		if r.ImageConcurrency == 0 {
			r.ImageConcurrency = current.ImageConcurrency
		}
		if r.VideoConcurrency == 0 {
			r.VideoConcurrency = current.VideoConcurrency
		}
	}
	if r.ImageConcurrency < 1 || r.ImageConcurrency > 100 || r.VideoConcurrency < 1 || r.VideoConcurrency > 100 {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "任务并发参数无效"))
		return
	}
	v := &model.TaskConfig{MaxActiveTasks: r.MaxActiveTasks, PromptMaxLength: r.PromptMaxLength, MaxRetryAttempts: r.MaxRetryAttempts, ImageConcurrency: r.ImageConcurrency, VideoConcurrency: r.VideoConcurrency}
	if e := h.configRepo.UpdateTaskConfig(v); e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_task_config", "task_config", 1, "更新任务配置", v)
	response.Success(c, v)
}
func (h *AdminConfigHandler) GetPromptOptimizationConfig(c *gin.Context) {
	v, e := h.configRepo.GetPromptOptimizationConfig()
	if e != nil {
		response.Error(c, e)
		return
	}
	response.Success(c, v)
}
func (h *AdminConfigHandler) SavePromptOptimizationConfig(c *gin.Context) {
	var r promptOptimizationConfigRequest
	if c.ShouldBindJSON(&r) != nil || r.MaxInputLength < 1 || r.MaxInputLength > 50000 || r.CreditCost < 0 || r.CreditCost > 100000 || r.RateLimitPerMinute < 1 || r.RateLimitPerMinute > 120 || (r.IsActive && r.ModelName == "") {
		response.Error(c, errors.New(errors.ErrCodeBadRequest, "提示词优化配置参数无效"))
		return
	}
	v := &model.PromptOptimizationConfig{IsActive: r.IsActive, ModelName: r.ModelName, SystemPrompt: r.SystemPrompt, MaxInputLength: r.MaxInputLength, CreditCost: r.CreditCost, RateLimitPerMinute: r.RateLimitPerMinute}
	if e := h.configRepo.UpdatePromptOptimizationConfig(v); e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_prompt_optimization_config", "prompt_optimization_config", 1, "更新提示词优化配置", v)
	response.Success(c, v)
}
func (h *AdminConfigHandler) GetModels(c *gin.Context) {
	v, e := h.modelRepo.GetAllModels()
	if e != nil {
		response.Error(c, e)
		return
	}
	response.Success(c, v)
}
func (h *AdminConfigHandler) UpdateModel(c *gin.Context) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var r modelUpdateRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	params, e := json.Marshal(r.ParamsConfig)
	if e != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	v, e := h.modelRepo.FindByID(id)
	if e != nil {
		response.Error(c, e)
		return
	}
	v.DisplayName = r.DisplayName
	v.Cost = r.Cost
	v.Description = r.Description
	v.ParamsConfig = params
	if e = h.modelRepo.Update(v); e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_model", "model", v.ID, "更新模型配置", v)
	response.Success(c, v)
}
func (h *AdminConfigHandler) UpdateModelStatus(c *gin.Context) {
	id, e := strconv.ParseInt(c.Param("id"), 10, 64)
	if e != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var r statusRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	if e = h.modelRepo.UpdateStatus(id, r.IsActive); e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_model_status", "model", id, "更新模型状态", r)
	response.Success(c, nil)
}

func (h *AdminConfigHandler) DeleteUnreferencedModels(c *gin.Context) {
	count, err := h.modelRepo.DeleteUnreferenced()
	if err != nil {
		response.Error(c, err)
		return
	}
	h.recordAudit(c, "delete_unreferenced_models", "model", 0, "清理未引用模型", map[string]int64{"count": count})
	response.Success(c, gin.H{"count": count})
}
func parseID(c *gin.Context) (int64, error) { return strconv.ParseInt(c.Param("id"), 10, 64) }
func (h *AdminConfigHandler) GetStorage(c *gin.Context) {
	v, e := h.storageService.GetStorageConfigs()
	if e != nil {
		response.Error(c, e)
		return
	}
	response.Success(c, v)
}
func (h *AdminConfigHandler) CreateStorage(c *gin.Context) {
	var r dto.StorageConfigRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	v, e := h.storageService.CreateStorageConfig(&r)
	if e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "create_storage", "storage", v.ID, "创建存储配置", v)
	response.Success(c, v)
}
func (h *AdminConfigHandler) UpdateStorage(c *gin.Context) {
	id, e := parseID(c)
	if e != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	var r dto.StorageConfigRequest
	if c.ShouldBindJSON(&r) != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	v, e := h.storageService.UpdateStorageConfig(id, &r)
	if e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "update_storage", "storage", id, "更新存储配置", v)
	response.Success(c, v)
}
func (h *AdminConfigHandler) EnableStorage(c *gin.Context) {
	id, e := parseID(c)
	if e == nil {
		e = h.storageService.EnableStorageConfig(id)
	}
	if e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "enable_storage", "storage", id, "启用存储配置", nil)
	response.Success(c, nil)
}
func (h *AdminConfigHandler) DeleteStorage(c *gin.Context) {
	id, e := parseID(c)
	if e == nil {
		e = h.storageService.DeleteStorageConfig(id)
	}
	if e != nil {
		response.Error(c, e)
		return
	}
	h.recordAudit(c, "delete_storage", "storage", id, "删除存储配置", nil)
	response.Success(c, nil)
}

func (h *AdminConfigHandler) GetResourcePolicies(c *gin.Context) {
	policies, err := h.resourcePolicyService.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, policies)
}

func (h *AdminConfigHandler) UpdateResourcePolicy(c *gin.Context) {
	var req resourcePolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errors.ErrBadRequest)
		return
	}
	policy, err := h.resourcePolicyService.Update(c.Param("type"), &model.ResourcePolicy{KeyPrefix: req.KeyPrefix, RetentionDays: req.RetentionDays, IsPublic: req.IsPublic, CacheMaxAge: req.CacheMaxAge, MaxSizeMB: req.MaxSizeMB})
	if err != nil {
		response.Error(c, err)
		return
	}
	h.recordAudit(c, "update_resource_policy", "resource_policy", 0, "更新资源策略", policy)
	response.Success(c, policy)
}
