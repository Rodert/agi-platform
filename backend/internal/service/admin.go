package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const adminLogDateLayout = "2006-01-02 15:04:05"

type AdminService struct {
	adminRepo  *repository.AdminRepository
	workRepo   *repository.WorkRepository
	taskRepo   *repository.TaskRepository
	userRepo   *repository.UserRepository
	creditRepo *repository.CreditRepository
	assetRepo  *repository.MediaAssetRepository
	storage    *objectstorage.Manager
	jwtConfig  *config.JWTConfig
	db         *gorm.DB
}

func NewAdminService(
	adminRepo *repository.AdminRepository,
	workRepo *repository.WorkRepository,
	taskRepo *repository.TaskRepository,
	userRepo *repository.UserRepository,
	creditRepo *repository.CreditRepository,
	assetRepo *repository.MediaAssetRepository,
	storage *objectstorage.Manager,
	jwtConfig *config.JWTConfig,
	db *gorm.DB,
) *AdminService {
	return &AdminService{
		adminRepo:  adminRepo,
		workRepo:   workRepo,
		taskRepo:   taskRepo,
		userRepo:   userRepo,
		creditRepo: creditRepo,
		assetRepo:  assetRepo,
		storage:    storage,
		jwtConfig:  jwtConfig,
		db:         db,
	}
}

// Login 管理员登录
func (s *AdminService) Login(req *dto.AdminLoginRequest, ip string) (*dto.AdminLoginResponse, error) {
	// 1. 查找管理员
	admin, err := s.adminRepo.FindByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.recordLoginFailure(ip)
			return nil, errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误")
		}
		return nil, err
	}

	// 2. 验证密码
	if !utils.CheckPassword(req.Password, admin.PasswordHash) {
		s.recordLoginFailure(ip)
		return nil, errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误")
	}

	// 3. 更新登录信息
	if err := s.adminRepo.UpdateLastLogin(admin.ID, ip); err != nil {
		return nil, err
	}

	// 4. 生成Token（管理员）
	token, err := jwt.GenerateAdminToken(admin.ID, admin.Username, admin.Role, s.jwtConfig)
	if err != nil {
		return nil, err
	}

	// 5. 记录登录日志
	s.recordAudit(&model.AdminLog{
		AdminID:     admin.ID,
		Action:      "login",
		Description: "管理员登录",
		IP:          ip,
		CreatedAt:   time.Now(),
	})

	return &dto.AdminLoginResponse{
		Token: token,
		Admin: &dto.AdminInfo{
			ID:       admin.ID,
			Username: admin.Username,
			Name:     admin.Name,
			Role:     admin.Role,
		},
	}, nil
}

func (s *AdminService) GetProfile(adminID int64) (*dto.AdminProfileResponse, error) {
	admin, err := s.adminRepo.FindByID(adminID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return nil, errors.ErrUnauthorized }
		return nil, err
	}
	profile := &dto.AdminProfileResponse{ID: admin.ID, Username: admin.Username, Name: admin.Name, Role: admin.Role, LastLoginIP: admin.LastLoginIP, CreatedAt: admin.CreatedAt.Format(adminLogDateLayout)}
	if admin.LastLoginAt != nil { profile.LastLoginAt = admin.LastLoginAt.Format(adminLogDateLayout) }
	return profile, nil
}

func (s *AdminService) UpdateProfile(adminID int64, req *dto.AdminUpdateProfileRequest) (*dto.AdminProfileResponse, error) {
	admin, err := s.adminRepo.FindByID(adminID)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return nil, errors.ErrUnauthorized }
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if name == "" { return nil, errors.New(errors.ErrCodeBadRequest, "姓名不能为空") }
	if (req.CurrentPassword == "") != (req.NewPassword == "") { return nil, errors.New(errors.ErrCodeBadRequest, "修改密码需同时填写当前密码和新密码") }
	passwordHash := ""
	passwordChanged := req.NewPassword != ""
	if passwordChanged {
		if !utils.CheckPassword(req.CurrentPassword, admin.PasswordHash) { return nil, errors.New(errors.ErrCodeBadRequest, "当前密码不正确") }
		passwordHash, err = utils.HashPassword(req.NewPassword)
		if err != nil { return nil, err }
	}
	before, _ := json.Marshal(map[string]interface{}{"name": admin.Name})
	if err := s.adminRepo.UpdateProfile(adminID, name, passwordHash); err != nil { return nil, err }
	admin.Name = name
	after, _ := json.Marshal(map[string]interface{}{"name": name, "password_changed": passwordChanged})
	s.recordAudit(&model.AdminLog{AdminID: adminID, Action: "update_profile", TargetType: "admin", TargetID: adminID, BeforeData: string(before), AfterData: string(after), Description: "更新个人资料", CreatedAt: time.Now()})
	return s.GetProfile(adminID)
}

// recordLoginFailure must never change the authentication result. AdminID 0
// represents an unauthenticated or unknown actor; admin_logs has no foreign key.
func (s *AdminService) recordLoginFailure(ip string) {
	s.recordAudit(&model.AdminLog{
		AdminID:     0,
		Action:      "login_failed",
		TargetType:  "admin_login",
		Description: "管理员登录失败",
		IP:          ip,
		CreatedAt:   time.Now(),
	})
}

// recordAudit makes audit persistence observable without turning an already
// completed administrative action into a failed response.
func (s *AdminService) recordAudit(log *model.AdminLog) {
	if err := s.adminRepo.CreateLog(log); err != nil {
		if logger.Log != nil {
			logger.Error("记录管理员审计日志失败", zap.String("action", log.Action), zap.Error(err))
		}
	}
}

// GetLogs lists persisted administrative audit records.
func (s *AdminService) GetLogs(req *dto.AdminLogListRequest) ([]*dto.AdminLogResponse, int64, error) {
	startAt, err := parseAdminLogTime(req.StartAt)
	if err != nil {
		return nil, 0, err
	}
	endAt, err := parseAdminLogTime(req.EndAt)
	if err != nil {
		return nil, 0, err
	}
	if startAt != nil && endAt != nil && endAt.Before(*startAt) {
		return nil, 0, errors.New(errors.ErrCodeBadRequest, "结束时间不能早于开始时间")
	}
	logs, total, err := s.adminRepo.ListLogs(req.Operator, req.Action, startAt, endAt, req.LoginOnly, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	// 转换为响应格式
	responses := make([]*dto.AdminLogResponse, len(logs))
	for i, log := range logs {
		response := &dto.AdminLogResponse{
			ID:          log.ID,
			AdminID:     log.AdminID,
			Action:      log.Action,
			TargetType:  log.TargetType,
			TargetID:    log.TargetID,
			BeforeData:  log.BeforeData,
			AfterData:   log.AfterData,
			Description: log.Description,
			IP:          log.IP,
			CreatedAt:   log.CreatedAt.Format(adminLogDateLayout),
		}
		if log.Admin != nil {
			response.Admin = &dto.AdminInfo{
				ID:       log.Admin.ID,
				Username: log.Admin.Username,
				Name:     log.Admin.Name,
				Role:     log.Admin.Role,
			}
		}
		responses[i] = response
	}

	return responses, total, nil
}

func (s *AdminService) ListDatabaseTables() ([]dto.DatabaseTableResponse, error) {
	tables, err := s.adminRepo.ListDatabaseTables()
	if err != nil { return nil, err }
	result := make([]dto.DatabaseTableResponse, len(tables))
	for index, table := range tables {
		result[index] = dto.DatabaseTableResponse{Name: table.Name, Comment: table.Comment}
	}
	return result, nil
}

func (s *AdminService) GetDatabaseTable(adminID int64, tableName string, page, pageSize int) (*dto.DatabaseTableDataResponse, error) {
	if page < 1 { page = 1 }
	if pageSize < 1 { pageSize = 20 }
	if pageSize > 100 { pageSize = 100 }
	columns, rows, total, err := s.adminRepo.GetDatabaseTable(tableName, page, pageSize)
	if err != nil {
		if err == gorm.ErrRecordNotFound { return nil, errors.New(errors.ErrCodeNotFound, "数据表不存在") }
		return nil, err
	}
	resultColumns := make([]dto.DatabaseColumnResponse, len(columns))
	for index, column := range columns {
		resultColumns[index] = dto.DatabaseColumnResponse{Name: column.Name, Type: column.Type, Nullable: column.Nullable, PrimaryKey: column.PrimaryKey}
	}
	s.recordAudit(&model.AdminLog{AdminID: adminID, Action: "browse_database", TargetType: "database_table", Description: fmt.Sprintf("浏览数据表 %s，第 %d 页", tableName, page), CreatedAt: time.Now()})
	return &dto.DatabaseTableDataResponse{Table: tableName, Columns: resultColumns, Rows: rows, Total: total, Page: page, PageSize: pageSize}, nil
}

func parseAdminLogTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.ParseInLocation(adminLogDateLayout, value, time.Local)
	if err != nil {
		return nil, errors.New(errors.ErrCodeBadRequest, fmt.Sprintf("时间格式必须为 %s", adminLogDateLayout))
	}
	return &parsed, nil
}

// AuditWork 审核作品
func (s *AdminService) AuditWork(adminID, workID int64, req *dto.AdminWorkAuditRequest) error {
	// 1. 获取作品
	work, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrWorkNotFound
		}
		return err
	}

	// 2. 检查审核状态
	if work.AuditStatus != "pending" {
		return errors.New(errors.ErrCodeBadRequest, "该作品已审核")
	}

	// 3. An approved work gets an independent, permanent object. Rejected works
	// keep using the temporary task object until its normal lifecycle removes it.
	var publishedAssets []*model.MediaAsset
	if req.Status == "approved" {
		publishedAssets, err = s.promoteWorkAssets(work)
		if err != nil {
			return errors.New(errors.ErrCodeBadRequest, "无法发布作品资源: "+err.Error())
		}
	}

	// 4. 更新作品状态和审核记录。对象已经复制成功后才允许作品进入首页列表。
	now := time.Now()
	work.AuditStatus = req.Status
	work.AuditReason = req.Reason
	work.AuditAdminID = adminID
	work.AuditedAt = &now
	work.UpdatedAt = now
	if req.Status == "approved" {
		work.PublishedAt = &now
	}
	audit := &model.WorkAudit{
		WorkID:    workID,
		AdminID:   adminID,
		Status:    req.Status,
		Reason:    req.Reason,
		AuditedAt: now,
		CreatedAt: now,
	}
	log := &model.AdminLog{
		AdminID:     adminID,
		Action:      "audit_work",
		TargetType:  "work",
		TargetID:    workID,
		Description: "审核作品：" + req.Status,
		CreatedAt:   time.Now(),
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(work).Error; err != nil {
			return err
		}
		for _, asset := range publishedAssets {
			if err := tx.Create(asset).Error; err != nil {
				return err
			}
		}
		if err := tx.Create(audit).Error; err != nil {
			return err
		}
		return tx.Create(log).Error
	})
}

// UpdateWorkPublicationStatus changes whether an approved work is visible in the
// public home feed. It deliberately does not copy or delete its permanent assets.
func (s *AdminService) UpdateWorkPublicationStatus(adminID, workID int64, req *dto.AdminWorkStatusRequest) error {
	work, err := s.workRepo.FindByID(workID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.ErrWorkNotFound
		}
		return err
	}
	if req.Status == work.AuditStatus {
		return errors.New(errors.ErrCodeBadRequest, "作品已处于该状态")
	}
	if req.Status == "offline" && work.AuditStatus != "approved" {
		return errors.New(errors.ErrCodeBadRequest, "只有已上架作品可以下架")
	}
	if req.Status == "approved" && work.AuditStatus != "offline" {
		return errors.New(errors.ErrCodeBadRequest, "只有已下架作品可以重新上架")
	}

	now := time.Now()
	work.AuditStatus = req.Status
	work.AuditReason = req.Reason
	work.AuditAdminID = adminID
	work.AuditedAt = &now
	work.UpdatedAt = now
	audit := &model.WorkAudit{WorkID: workID, AdminID: adminID, Status: req.Status, Reason: req.Reason, AuditedAt: now, CreatedAt: now}
	action, description := "offline_work", "下架作品"
	if req.Status == "approved" {
		action, description = "republish_work", "重新上架作品"
	}
	log := &model.AdminLog{AdminID: adminID, Action: action, TargetType: "work", TargetID: workID, Description: description, CreatedAt: now}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(work).Error; err != nil {
			return err
		}
		if err := tx.Create(audit).Error; err != nil {
			return err
		}
		return tx.Create(log).Error
	})
}

func (s *AdminService) promoteWorkAssets(work *model.Work) ([]*model.MediaAsset, error) {
	task, err := s.taskRepo.FindByID(work.TaskID)
	if err != nil {
		return nil, err
	}
	if task.Status != "success" || task.ResultURL == "" {
		return nil, errors.New(errors.ErrCodeBadRequest, "原始生成任务不可用于发布")
	}
	assetType, publishedType := "image", "published_image"
	if work.Type == "video" {
		assetType, publishedType = "video", "published_video"
	}
	content, err := s.assetRepo.FindByTaskAndURL(task.ID, assetType, task.ResultURL)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "原始生成资源不存在或已清理")
		}
		return nil, err
	}
	published, err := s.storage.Promote(context.Background(), content, publishedType)
	if err != nil {
		return nil, err
	}
	contentAsset := mediaAssetFromPublished(work, task.ID, published)
	assets := []*model.MediaAsset{contentAsset}
	if work.Type == "image" {
		work.ImageURL, work.VideoURL = published.PublicURL, ""
		return assets, nil
	}

	work.ImageURL, work.VideoURL = "", published.PublicURL
	if task.ThumbnailURL == "" || task.ThumbnailURL == task.ResultURL {
		return assets, nil
	}
	thumbnail, err := s.assetRepo.FindByTaskAndURL(task.ID, "thumbnail", task.ThumbnailURL)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "视频缩略图不存在或已清理")
		}
		return nil, err
	}
	publishedThumbnail, err := s.storage.Promote(context.Background(), thumbnail, "published_thumbnail")
	if err != nil {
		return nil, err
	}
	work.ImageURL = publishedThumbnail.PublicURL
	return append(assets, mediaAssetFromPublished(work, task.ID, publishedThumbnail)), nil
}

func mediaAssetFromPublished(work *model.Work, taskID int64, stored *objectstorage.StoredObject) *model.MediaAsset {
	return &model.MediaAsset{TaskID: &taskID, UserID: work.UserID, StorageConfigID: stored.StorageConfigID, ResourceType: stored.ResourceType, ObjectKey: stored.ObjectKey, PublicURL: stored.PublicURL, ContentType: stored.ContentType, SizeBytes: stored.SizeBytes, ExpiresAt: stored.ExpiresAt, CreatedAt: time.Now()}
}

// GetPendingWorks 获取待审核作品列表
func (s *AdminService) GetPendingWorks(page, pageSize int) ([]*dto.WorkResponse, int64, error) {
	return s.GetWorks("pending", page, pageSize)
}

// GetWorks lists all works for administrators, optionally filtered by lifecycle status.
func (s *AdminService) GetWorks(status string, page, pageSize int) ([]*dto.WorkResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	works, total, err := s.workRepo.FindAdminList(status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.WorkResponse, len(works))
	for i, work := range works {
		responses[i] = &dto.WorkResponse{
			ID:          work.ID,
			UserID:      work.UserID,
			Title:       work.Title,
			Prompt:      work.Prompt,
			Category:    work.Category,
			Type:        work.Type,
			ImageURL:    work.ImageURL,
			VideoURL:    work.VideoURL,
			AuditStatus: work.AuditStatus,
			CreatedAt:   work.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return responses, total, nil
}

// GetStats 获取统计数据
func (s *AdminService) GetStats() (*dto.AdminStatsResponse, error) {
	stats, err := s.adminRepo.GetStats()
	if err != nil {
		return nil, err
	}

	return &dto.AdminStatsResponse{
		TotalUsers:   stats["total_users"],
		TotalTasks:   stats["total_tasks"],
		TotalWorks:   stats["total_works"],
		PendingWorks: stats["pending_works"],
		TodayUsers:   stats["today_users"],
		TodayTasks:   stats["today_tasks"],
		TodayWorks:   stats["today_works"],
	}, nil
}

func (s *AdminService) GetReport(start, end time.Time) (*dto.AdminReportResponse, error) {
	records, err := s.adminRepo.GetReportRecords(start, end)
	if err != nil {
		return nil, err
	}
	result := &dto.AdminReportResponse{
		StartDate: start.Format("2006-01-02"),
		EndDate:   end.Add(-time.Nanosecond).Format("2006-01-02"),
	}
	daily := make(map[string]*dto.AdminReportDailyPoint)
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		date := day.Format("2006-01-02")
		daily[date] = &dto.AdminReportDailyPoint{Date: date}
	}
	taskTypes, taskModels, taskChannels := map[string]int64{}, map[string]int64{}, map[string]int64{}
	workStatuses, workCategories := map[string]int64{}, map[string]int64{}
	activeUsers := make(map[int64]struct{})
	dailyActiveUsers := make(map[string]map[int64]struct{})

	for _, user := range records.Users {
		result.Summary.NewUsers++
		daily[user.CreatedAt.Format("2006-01-02")].NewUsers++
	}
	for _, task := range records.Tasks {
		point := daily[task.CreatedAt.Format("2006-01-02")]
		result.Summary.Tasks++
		point.Tasks++
		activeUsers[task.UserID] = struct{}{}
		date := task.CreatedAt.Format("2006-01-02")
		if dailyActiveUsers[date] == nil {
			dailyActiveUsers[date] = make(map[int64]struct{})
		}
		dailyActiveUsers[date][task.UserID] = struct{}{}
		switch task.Status {
		case "success":
			result.Summary.SuccessTasks++
			point.SuccessTasks++
		case "failed":
			result.Summary.FailedTasks++
			point.FailedTasks++
		default:
			result.Summary.PendingTasks++
		}
		taskTypes[reportLabel(task.Type)]++
		taskModels[reportLabel(task.ModelName)]++
		channelName := "未指定"
		if task.Channel != nil {
			channelName = reportLabel(task.Channel.Name)
		}
		taskChannels[channelName]++
	}
	for _, ledger := range records.Ledgers {
		result.Summary.CreditsConsumed += int64(ledger.Amount)
		daily[ledger.CreatedAt.Format("2006-01-02")].CreditsConsumed += int64(ledger.Amount)
	}
	for _, work := range records.Works {
		result.Summary.Works++
		daily[work.CreatedAt.Format("2006-01-02")].Works++
		switch work.AuditStatus {
		case "pending":
			result.Summary.PendingWorks++
		case "approved":
			result.Summary.ApprovedWorks++
		case "rejected":
			result.Summary.RejectedWorks++
		case "offline":
			result.Summary.OfflineWorks++
		}
		workStatuses[reportLabel(work.AuditStatus)]++
		workCategories[reportLabel(work.Category)]++
	}
	result.Summary.ActiveUsers = int64(len(activeUsers))
	for date, users := range dailyActiveUsers {
		daily[date].ActiveUsers = int64(len(users))
	}
	if result.Summary.Tasks > 0 {
		result.Summary.SuccessRate = float64(result.Summary.SuccessTasks) * 100 / float64(result.Summary.Tasks)
	}
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		result.Daily = append(result.Daily, *daily[day.Format("2006-01-02")])
	}
	result.TaskTypes = reportBreakdown(taskTypes)
	result.TaskModels = reportBreakdown(taskModels)
	result.TaskChannels = reportBreakdown(taskChannels)
	result.WorkStatuses = reportBreakdown(workStatuses)
	result.WorkCategories = reportBreakdown(workCategories)
	return result, nil
}

func reportLabel(value string) string {
	if value = strings.TrimSpace(value); value == "" {
		return "未分类"
	}
	return value
}

func reportBreakdown(counts map[string]int64) []dto.AdminReportBreakdownItem {
	items := make([]dto.AdminReportBreakdownItem, 0, len(counts))
	for name, value := range counts {
		items = append(items, dto.AdminReportBreakdownItem{Name: name, Value: value})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Value == items[j].Value {
			return items[i].Name < items[j].Name
		}
		return items[i].Value > items[j].Value
	})
	return items
}

func (s *AdminService) GetTaskList(req *dto.AdminTaskListRequest) ([]*dto.AdminTaskResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	tasks, total, err := s.taskRepo.FindAdminTasks(req.Keyword, req.Status, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}
	responses := make([]*dto.AdminTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		params := map[string]interface{}{}
		if task.Request != nil && len(task.Request.Params) > 0 {
			_ = json.Unmarshal(task.Request.Params, &params)
		}
		providerResponse := map[string]interface{}{}
		if len(task.ProviderResponse) > 0 {
			_ = json.Unmarshal(task.ProviderResponse, &providerResponse)
		}
		item := &dto.AdminTaskResponse{ID: task.ID, UserID: task.UserID, ProviderTaskID: task.ProviderTaskID, ProviderStatus: task.ProviderStatus, ProviderResponse: providerResponse, ModelName: task.ModelName, Type: task.Type, Status: task.Status, Progress: task.Progress, Prompt: task.Prompt, Params: params, Cost: task.Cost, ResultURL: task.ResultURL, ThumbnailURL: task.ThumbnailURL, ErrorMsg: task.ErrorMsg, AttemptCount: task.AttemptCount, MaxRetryAttempts: task.MaxRetryAttempts, CreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05")}
		if task.User != nil {
			item.UserName, item.UserEmail = task.User.Name, task.User.Email
		}
		if task.Channel != nil {
			item.ChannelName = task.Channel.Name
		}
		if task.CompletedAt != nil {
			item.CompletedAt = task.CompletedAt.Format("2006-01-02 15:04:05")
		}
		if task.LastRetryAt != nil {
			item.LastRetryAt = task.LastRetryAt.Format("2006-01-02 15:04:05")
		}
		if task.LastPolledAt != nil {
			item.LastPolledAt = task.LastPolledAt.Format("2006-01-02 15:04:05")
		}
		for _, attempt := range task.Attempts {
			entry := dto.AdminTaskAttemptResponse{Attempt: attempt.Attempt, Status: attempt.Status, ErrorMsg: attempt.ErrorMsg, StartedAt: attempt.StartedAt.Format("2006-01-02 15:04:05")}
			if attempt.CompletedAt != nil {
				entry.CompletedAt = attempt.CompletedAt.Format("2006-01-02 15:04:05")
			}
			item.Attempts = append(item.Attempts, entry)
		}
		for _, asset := range task.Assets {
			entry := dto.AdminMediaAssetResponse{ResourceType: asset.ResourceType, StorageConfigID: asset.StorageConfigID, ObjectKey: asset.ObjectKey, PublicURL: asset.PublicURL, ContentType: asset.ContentType, SizeBytes: asset.SizeBytes}
			if asset.ExpiresAt != nil {
				entry.ExpiresAt = asset.ExpiresAt.Format("2006-01-02 15:04:05")
			}
			item.Assets = append(item.Assets, entry)
		}
		responses = append(responses, item)
	}
	return responses, total, nil
}
