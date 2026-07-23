package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/response"
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
}

func NewAdminConfigHandler(c *repository.ConfigRepository, m *repository.AIModelRepository, p *repository.AIProviderAccountRepository, cm *repository.ChannelModelRepository, catalog *service.ChannelCatalogService, s *service.StorageConfigService, rp *service.ResourcePolicyService, e *service.EmailService) *AdminConfigHandler {
	return &AdminConfigHandler{configRepo: c, modelRepo: m, providerRepo: p, channelModelRepo: cm, channelCatalogService: catalog, storageService: s, resourcePolicyService: rp, emailService: e}
}

type basicConfigRequest struct {
	SiteName        string `json:"site_name"`
	SiteDesc        string `json:"site_desc"`
	RegisterEnabled bool   `json:"register_enabled"`
	RegisterCredits int    `json:"register_credits"`
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
	keys := []string{"site_name", "site_desc", "register_enabled", "new_user_gift_amount"}
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
	values := []struct{ k, v, t, d string }{{"site_name", r.SiteName, "string", "网站名称"}, {"site_desc", r.SiteDesc, "string", "网站描述"}, {"register_enabled", strconv.FormatBool(r.RegisterEnabled), "bool", "注册开关"}, {"new_user_gift_amount", strconv.Itoa(r.RegisterCredits), "int", "新用户赠送积分"}}
	for _, v := range values {
		if e := h.configRepo.UpsertSystemConfig(v.k, v.v, v.t, "basic", v.d); e != nil {
			response.Error(c, e)
			return
		}
	}
	response.Success(c, nil)
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
	v := &model.TaskConfig{MaxActiveTasks: r.MaxActiveTasks, PromptMaxLength: r.PromptMaxLength, MaxRetryAttempts: r.MaxRetryAttempts}
	if e := h.configRepo.UpdateTaskConfig(v); e != nil {
		response.Error(c, e)
		return
	}
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
	response.Success(c, nil)
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
	response.Success(c, policy)
}
