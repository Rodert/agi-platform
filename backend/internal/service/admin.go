package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/jwt"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

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
			return nil, errors.New(errors.ErrCodeUnauthorized, "用户名或密码错误")
		}
		return nil, err
	}

	// 2. 验证密码
	if !utils.CheckPassword(req.Password, admin.PasswordHash) {
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
	if err := s.adminRepo.CreateLog(&model.AdminLog{
		AdminID:     admin.ID,
		Action:      "login",
		Description: "管理员登录",
		IP:          ip,
		CreatedAt:   time.Now(),
	}); err != nil {
		return nil, err
	}

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
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	works, total, err := s.workRepo.FindPendingWorks(page, pageSize)
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
		item := &dto.AdminTaskResponse{ID: task.ID, UserID: task.UserID, ModelName: task.ModelName, Type: task.Type, Status: task.Status, Progress: task.Progress, Prompt: task.Prompt, Params: params, Cost: task.Cost, ResultURL: task.ResultURL, ThumbnailURL: task.ThumbnailURL, ErrorMsg: task.ErrorMsg, AttemptCount: task.AttemptCount, MaxRetryAttempts: task.MaxRetryAttempts, CreatedAt: task.CreatedAt.Format("2006-01-02 15:04:05")}
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
