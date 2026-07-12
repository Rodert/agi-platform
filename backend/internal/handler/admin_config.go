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
	configRepo *repository.ConfigRepository
	modelRepo *repository.AIModelRepository
	providerRepo *repository.AIProviderAccountRepository
	storageService *service.StorageConfigService
	emailService *service.EmailService
}

func NewAdminConfigHandler(c *repository.ConfigRepository, m *repository.AIModelRepository, p *repository.AIProviderAccountRepository, s *service.StorageConfigService, e *service.EmailService) *AdminConfigHandler {
	return &AdminConfigHandler{configRepo:c, modelRepo:m, providerRepo:p, storageService:s, emailService:e}
}

type basicConfigRequest struct { SiteName string `json:"site_name"`; SiteDesc string `json:"site_desc"`; RegisterEnabled bool `json:"register_enabled"`; RegisterCredits int `json:"register_credits"` }
type emailConfigRequest struct { SMTPHost string `json:"smtp_host"`; SMTPPort int `json:"smtp_port"`; SMTPUser string `json:"smtp_user"`; SMTPPassword string `json:"smtp_password"`; SMTPSSL bool `json:"smtp_ssl"`; FromName string `json:"from_name"`; FromEmail string `json:"from_email"`; IsActive bool `json:"is_active"` }
type modelUpdateRequest struct { DisplayName string `json:"display_name"`; Cost int `json:"cost"`; Description string `json:"description"`; ProviderAccountID *int64 `json:"provider_account_id"`; ParamsConfig map[string]interface{} `json:"params_config"` }
type statusRequest struct { IsActive bool `json:"is_active"` }
type testEmailRequest struct { Email string `json:"email" binding:"required,email"` }

func (h *AdminConfigHandler) GetBasic(c *gin.Context) { keys:=[]string{"site_name","site_desc","register_enabled","new_user_gift_amount"}; out:=map[string]string{}; for _,k:=range keys { if v,e:=h.configRepo.GetSystemConfig(k); e==nil { out[k]=v.Value } }; response.Success(c,out) }
func (h *AdminConfigHandler) SaveBasic(c *gin.Context) { var r basicConfigRequest; if c.ShouldBindJSON(&r)!=nil { response.Error(c,errors.ErrBadRequest); return }; values:=[]struct{k,v,t,d string}{{"site_name",r.SiteName,"string","网站名称"},{"site_desc",r.SiteDesc,"string","网站描述"},{"register_enabled",strconv.FormatBool(r.RegisterEnabled),"bool","注册开关"},{"new_user_gift_amount",strconv.Itoa(r.RegisterCredits),"int","新用户赠送积分"}}; for _,v:=range values { if e:=h.configRepo.UpsertSystemConfig(v.k,v.v,v.t,"basic",v.d);e!=nil {response.Error(c,e);return} }; response.Success(c,nil) }
func (h *AdminConfigHandler) GetEmail(c *gin.Context) { v,e:=h.configRepo.GetEmailConfig();if e!=nil{response.Error(c,e);return}; v.SMTPPassword="";response.Success(c,v) }
func (h *AdminConfigHandler) SaveEmail(c *gin.Context) { var r emailConfigRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return}; old,e:=h.configRepo.GetEmailConfig();if e!=nil{old=&model.EmailConfig{ID:1}};if r.SMTPPassword==""{r.SMTPPassword=old.SMTPPassword}; v:=&model.EmailConfig{ID:1,SMTPHost:r.SMTPHost,SMTPPort:r.SMTPPort,SMTPUser:r.SMTPUser,SMTPPassword:r.SMTPPassword,SMTPSSL:r.SMTPSSL,FromName:r.FromName,FromEmail:r.FromEmail,IsActive:r.IsActive};if e=h.configRepo.UpdateEmailConfig(v);e!=nil{response.Error(c,e);return};response.Success(c,nil) }
func (h *AdminConfigHandler) TestEmail(c *gin.Context) { var r testEmailRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return};if e:=h.emailService.SendVerificationCode(r.Email,"123456");e!=nil{response.Error(c,e);return};response.Success(c,nil) }
func (h *AdminConfigHandler) GetModels(c *gin.Context) { v,e:=h.modelRepo.GetAllModels();if e!=nil{response.Error(c,e);return};for _,item:=range v{if item.ProviderAccount!=nil&&item.ProviderAccount.APIKey!=""{item.ProviderAccount.APIKey="******"}};response.Success(c,v) }
func (h *AdminConfigHandler) UpdateModel(c *gin.Context) { id,e:=strconv.ParseInt(c.Param("id"),10,64);if e!=nil{response.Error(c,errors.ErrBadRequest);return};var r modelUpdateRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return};params,e:=json.Marshal(r.ParamsConfig);if e!=nil{response.Error(c,errors.ErrBadRequest);return};v,e:=h.modelRepo.FindByID(id);if e!=nil{response.Error(c,e);return};if r.ProviderAccountID!=nil{account,findErr:=h.providerRepo.Find(*r.ProviderAccountID);if findErr!=nil{response.Error(c,findErr);return};v.Provider=account.Provider};v.DisplayName=r.DisplayName;v.Cost=r.Cost;v.Description=r.Description;v.ProviderAccountID=r.ProviderAccountID;v.ParamsConfig=params;if e=h.modelRepo.Update(v);e!=nil{response.Error(c,e);return};response.Success(c,v) }
func (h *AdminConfigHandler) UpdateModelStatus(c *gin.Context) { id,e:=strconv.ParseInt(c.Param("id"),10,64);if e!=nil{response.Error(c,errors.ErrBadRequest);return};var r statusRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return};if e=h.modelRepo.UpdateStatus(id,r.IsActive);e!=nil{response.Error(c,e);return};response.Success(c,nil) }
func parseID(c *gin.Context)(int64,error){return strconv.ParseInt(c.Param("id"),10,64)}
func (h *AdminConfigHandler) GetStorage(c *gin.Context){v,e:=h.storageService.GetStorageConfigs();if e!=nil{response.Error(c,e);return};response.Success(c,v)}
func (h *AdminConfigHandler) CreateStorage(c *gin.Context){var r dto.StorageConfigRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return};v,e:=h.storageService.CreateStorageConfig(&r);if e!=nil{response.Error(c,e);return};response.Success(c,v)}
func (h *AdminConfigHandler) UpdateStorage(c *gin.Context){id,e:=parseID(c);if e!=nil{response.Error(c,errors.ErrBadRequest);return};var r dto.StorageConfigRequest;if c.ShouldBindJSON(&r)!=nil{response.Error(c,errors.ErrBadRequest);return};v,e:=h.storageService.UpdateStorageConfig(id,&r);if e!=nil{response.Error(c,e);return};response.Success(c,v)}
func (h *AdminConfigHandler) EnableStorage(c *gin.Context){id,e:=parseID(c);if e==nil{e=h.storageService.EnableStorageConfig(id)};if e!=nil{response.Error(c,e);return};response.Success(c,nil)}
func (h *AdminConfigHandler) DeleteStorage(c *gin.Context){id,e:=parseID(c);if e==nil{e=h.storageService.DeleteStorageConfig(id)};if e!=nil{response.Error(c,e);return};response.Success(c,nil)}
