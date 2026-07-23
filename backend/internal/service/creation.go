package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/queue"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreationService struct {
	taskRepo         *repository.TaskRepository
	requestRepo      *repository.GenerationRequestRepository
	aiModelRepo      *repository.AIModelRepository
	channelModelRepo *repository.ChannelModelRepository
	creditRepo       *repository.CreditRepository
	assetRepo        *repository.MediaAssetRepository
	configRepo       *repository.ConfigRepository
	storageService   *StorageService
	queueProducer    *queue.Producer
	db               *gorm.DB
}

func NewCreationService(
	taskRepo *repository.TaskRepository,
	requestRepo *repository.GenerationRequestRepository,
	aiModelRepo *repository.AIModelRepository,
	channelModelRepo *repository.ChannelModelRepository,
	creditRepo *repository.CreditRepository,
	assetRepo *repository.MediaAssetRepository,
	configRepo *repository.ConfigRepository,
	storageService *StorageService,
	queueProducer *queue.Producer,
	db *gorm.DB,
) *CreationService {
	return &CreationService{
		taskRepo:         taskRepo,
		requestRepo:      requestRepo,
		aiModelRepo:      aiModelRepo,
		channelModelRepo: channelModelRepo,
		creditRepo:       creditRepo,
		assetRepo:        assetRepo,
		configRepo:       configRepo,
		storageService:   storageService,
		queueProducer:    queueProducer,
		db:               db,
	}
}

// CreateImageTask 创建图片生成任务
func (s *CreationService) CreateImageTask(userID int64, req *dto.CreateImageTaskRequest) (*dto.TaskResponse, error) {
	taskConfig, err := s.getTaskConfig()
	if err != nil {
		return nil, err
	}
	if err := validatePrompt(req.Prompt, taskConfig.PromptMaxLength); err != nil {
		return nil, err
	}

	// 1. 验证模型
	aiModel, err := s.aiModelRepo.FindByName(req.ModelName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "模型不存在")
		}
		return nil, err
	}

	if aiModel.Type != "image" {
		return nil, errors.New(errors.ErrCodeBadRequest, "该模型不支持图片生成")
	}
	channelBinding, err := s.channelModelRepo.SelectActiveChannel(aiModel.ID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeBadRequest, "该模型当前没有可用渠道")
	}

	// 2. 校验模型参数并计算费用
	params, cost, err := s.resolveModelParams(aiModel, req.Params)
	if err != nil {
		return nil, err
	}

	// 3. 检查并发任务数限制
	if err := s.checkConcurrentLimit(userID, taskConfig.MaxActiveTasks); err != nil {
		return nil, err
	}

	// 4. 处理参考图（Base64 → URL）
	var referenceURL string
	var referenceStored *objectstorage.StoredObject
	if req.ReferenceImage != "" {
		stored, err := s.storageService.UploadBase64Image(context.Background(), req.ReferenceImage)
		if err != nil {
			return nil, errors.NewWithDetails(errors.ErrCodeUploadFailed, "参考图上传失败", err.Error())
		}
		referenceURL = stored.PublicURL
		referenceStored = stored
	}

	// 5. 构建任务参数
	if referenceURL != "" {
		params["reference_image_url"] = referenceURL
	}

	paramsJSON, _ := json.Marshal(params)

	// 6. 开启事务
	var task *model.Task
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 扣除积分
		if err := s.deductCredit(tx, userID, cost, "图片生成"); err != nil {
			return err
		}

		// 创建生成请求
		genReq := &model.GenerationRequest{
			UserID:    userID,
			Prompt:    req.Prompt,
			ModelName: req.ModelName,
			Type:      "image",
			Params:    datatypes.JSON(paramsJSON),
			Cost:      cost,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(genReq).Error; err != nil {
			return err
		}

		// 创建任务
		task = &model.Task{
			UserID:           userID,
			RequestID:        genReq.ID,
			Title:            utils.TruncateString(req.Prompt, 50),
			Type:             "image",
			Status:           "queued",
			Progress:         0,
			Prompt:           req.Prompt,
			ModelName:        req.ModelName,
			ChannelID:        channelBinding.ChannelID,
			MaxRetryAttempts: taskConfig.MaxRetryAttempts,
			Cost:             cost,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		if req.ReferenceImage != "" {
			if err := s.assetRepo.CreateTx(tx, mediaAssetFromStored(userID, &task.ID, referenceStored)); err != nil {
				return err
			}
		}

		// 更新请求的 TaskID
		if err := tx.Model(genReq).Update("task_id", task.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 7. 投递到队列
	if s.queueProducer != nil {
		if err := s.publishTask(task, params); err != nil {
			// 记录错误但不影响返回
			fmt.Printf("投递任务到队列失败: %v\n", err)
		}
	}

	return s.taskToResponse(task), nil
}

// CreateVideoTask 创建视频生成任务
func (s *CreationService) CreateVideoTask(userID int64, req *dto.CreateVideoTaskRequest) (*dto.TaskResponse, error) {
	taskConfig, err := s.getTaskConfig()
	if err != nil {
		return nil, err
	}
	if err := validatePrompt(req.Prompt, taskConfig.PromptMaxLength); err != nil {
		return nil, err
	}

	// 1. 验证模型
	aiModel, err := s.aiModelRepo.FindByName(req.ModelName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "模型不存在")
		}
		return nil, err
	}

	if aiModel.Type != "video" {
		return nil, errors.New(errors.ErrCodeBadRequest, "该模型不支持视频生成")
	}
	channelBinding, err := s.channelModelRepo.SelectActiveChannel(aiModel.ID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeBadRequest, "该模型当前没有可用渠道")
	}

	// 2. 校验模型参数并计算费用
	params, cost, err := s.resolveModelParams(aiModel, req.Params)
	if err != nil {
		return nil, err
	}

	// 3. 检查并发任务数限制
	if err := s.checkConcurrentLimit(userID, taskConfig.MaxActiveTasks); err != nil {
		return nil, err
	}

	// 4. Reference uploads use their configured public resource policy so each
	// upstream provider can fetch the persisted image URL directly.
	referenceAssets := make([]*objectstorage.StoredObject, 0, len(req.ReferenceImages)+2)
	if req.FirstFrameURL != "" {
		url, stored, err := s.prepareVideoReference(context.Background(), req.FirstFrameURL)
		if err != nil {
			return nil, errors.NewWithDetails(errors.ErrCodeUploadFailed, "首帧参考图处理失败", err.Error())
		}
		params["first_frame_url"] = url
		if stored != nil {
			referenceAssets = append(referenceAssets, stored)
		}
	}
	if req.LastFrameURL != "" {
		url, stored, err := s.prepareVideoReference(context.Background(), req.LastFrameURL)
		if err != nil {
			return nil, errors.NewWithDetails(errors.ErrCodeUploadFailed, "尾帧参考图处理失败", err.Error())
		}
		params["last_frame_url"] = url
		if stored != nil {
			referenceAssets = append(referenceAssets, stored)
		}
	}
	referenceURLs := make([]string, 0, len(req.ReferenceImages))
	for _, source := range req.ReferenceImages {
		url, stored, err := s.prepareVideoReference(context.Background(), source)
		if err != nil {
			return nil, errors.NewWithDetails(errors.ErrCodeUploadFailed, "参考图处理失败", err.Error())
		}
		if url != "" {
			referenceURLs = append(referenceURLs, url)
		}
		if stored != nil {
			referenceAssets = append(referenceAssets, stored)
		}
	}
	if len(referenceURLs) > 0 {
		params["reference_image_urls"] = referenceURLs
	}

	paramsJSON, _ := json.Marshal(params)

	// 5. 开启事务创建任务
	var task *model.Task
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// 扣除积分
		if err := s.deductCredit(tx, userID, cost, "视频生成"); err != nil {
			return err
		}

		// 创建生成请求
		genReq := &model.GenerationRequest{
			UserID:    userID,
			Prompt:    req.Prompt,
			ModelName: req.ModelName,
			Type:      "video",
			Params:    datatypes.JSON(paramsJSON),
			Cost:      cost,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(genReq).Error; err != nil {
			return err
		}

		// 创建任务
		task = &model.Task{
			UserID:           userID,
			RequestID:        genReq.ID,
			Title:            utils.TruncateString(req.Prompt, 50),
			Type:             "video",
			Status:           "queued",
			Progress:         0,
			Prompt:           req.Prompt,
			ModelName:        req.ModelName,
			ChannelID:        channelBinding.ChannelID,
			MaxRetryAttempts: taskConfig.MaxRetryAttempts,
			Cost:             cost,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		if err := tx.Create(task).Error; err != nil {
			return err
		}
		for _, stored := range referenceAssets {
			if err := s.assetRepo.CreateTx(tx, mediaAssetFromStored(userID, &task.ID, stored)); err != nil {
				return err
			}
		}

		// 更新请求的 TaskID
		if err := tx.Model(genReq).Update("task_id", task.ID).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 投递到队列
	if s.queueProducer != nil {
		if err := s.publishTask(task, params); err != nil {
			fmt.Printf("投递任务到队列失败: %v\n", err)
		}
	}

	return s.taskToResponse(task), nil
}

// GetTask 获取任务详情
func (s *CreationService) GetTask(userID, taskID int64) (*dto.TaskResponse, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrTaskNotFound
		}
		return nil, err
	}

	// 验证权限
	if task.UserID != userID {
		return nil, errors.ErrForbidden
	}

	return s.taskToResponse(task), nil
}

// DownloadTask opens a completed task result after verifying that it belongs
// to the requesting user. The handler streams the reader as an attachment.
func (s *CreationService) DownloadTask(userID, taskID int64) (io.ReadCloser, *model.MediaAsset, error) {
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.ErrTaskNotFound
		}
		return nil, nil, err
	}
	if task.UserID != userID {
		return nil, nil, errors.ErrForbidden
	}
	if task.Status != "success" || task.ResultURL == "" {
		return nil, nil, errors.New(errors.ErrCodeBadRequest, "任务结果尚不可下载")
	}

	resourceType := "image"
	if task.Type == "video" {
		resourceType = "video"
	}
	asset, err := s.assetRepo.FindByTaskAndURL(task.ID, resourceType, task.ResultURL)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, errors.New(errors.ErrCodeNotFound, "生成资源不存在或已清理")
		}
		return nil, nil, err
	}
	if asset.ExpiresAt != nil && !asset.ExpiresAt.After(time.Now()) {
		return nil, nil, errors.New(errors.ErrCodeNotFound, "生成资源已过期")
	}
	reader, err := s.storageService.Download(context.Background(), asset)
	if err != nil {
		return nil, nil, err
	}
	return reader, asset, nil
}

// GetTaskList 获取任务列表
func (s *CreationService) GetTaskList(userID int64, req *dto.TaskListRequest) ([]*dto.TaskResponse, int64, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	tasks, total, err := s.taskRepo.FindByUserID(userID, req.Status, req.Type, req.Page, req.PageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*dto.TaskResponse, len(tasks))
	for i, task := range tasks {
		responses[i] = s.taskToResponse(task)
	}

	return responses, total, nil
}

// GetModels 获取模型列表
func (s *CreationService) GetModels(modelType string) ([]*dto.GetModelsResponse, error) {
	models, err := s.aiModelRepo.GetActiveModels(modelType)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.GetModelsResponse, len(models))
	for i, m := range models {
		var paramsConfig map[string]interface{}
		if len(m.ParamsConfig) > 0 {
			json.Unmarshal(m.ParamsConfig, &paramsConfig)
		}

		responses[i] = &dto.GetModelsResponse{
			ID:           m.ID,
			Name:         m.Name,
			DisplayName:  m.DisplayName,
			Type:         m.Type,
			Provider:     m.Provider,
			Description:  m.Description,
			LogoURL:      m.LogoURL,
			Tag:          m.Tag,
			Cost:         m.Cost,
			ParamsConfig: paramsConfig,
		}
	}

	return responses, nil
}

type modelParamOption struct {
	Value     string `json:"value"`
	ExtraCost int    `json:"extra_cost"`
}

type modelParamConfig struct {
	Type    string             `json:"type"`
	Default interface{}        `json:"default"`
	Options []modelParamOption `json:"options"`
}

// resolveModelParams applies defaults and rejects values unsupported by the selected model.
func (s *CreationService) resolveModelParams(aiModel *model.AIModel, input map[string]interface{}) (map[string]interface{}, int, error) {
	configs := map[string]modelParamConfig{}
	if len(aiModel.ParamsConfig) > 0 {
		if err := json.Unmarshal(aiModel.ParamsConfig, &configs); err != nil {
			return nil, 0, errors.New(errors.ErrCodeInternalServer, "模型参数配置无效")
		}
	}
	for key := range input {
		if _, ok := configs[key]; !ok {
			return nil, 0, errors.New(errors.ErrCodeBadRequest, "模型不支持参数: "+key)
		}
	}
	params := make(map[string]interface{}, len(configs))
	cost := aiModel.Cost
	for key, config := range configs {
		value, exists := input[key]
		if !exists {
			value = config.Default
		}
		switch config.Type {
		case "switch":
			if _, ok := value.(bool); !ok {
				return nil, 0, errors.New(errors.ErrCodeBadRequest, "参数类型错误: "+key)
			}
		case "select":
			selected, ok := value.(string)
			if !ok {
				return nil, 0, errors.New(errors.ErrCodeBadRequest, "参数类型错误: "+key)
			}
			valid := false
			for _, option := range config.Options {
				if option.Value == selected {
					valid = true
					cost += option.ExtraCost
					break
				}
			}
			if !valid {
				return nil, 0, errors.New(errors.ErrCodeBadRequest, "模型不支持参数值: "+key)
			}
		default:
			return nil, 0, errors.New(errors.ErrCodeInternalServer, "模型参数类型无效: "+key)
		}
		params[key] = value
	}
	return params, cost, nil
}

func (s *CreationService) getTaskConfig() (*model.TaskConfig, error) {
	config, err := s.configRepo.GetTaskConfig()
	if err != nil {
		return nil, err
	}
	if config.MaxActiveTasks < 1 || config.PromptMaxLength < 1 || config.MaxRetryAttempts < 0 {
		return nil, errors.New(errors.ErrCodeInternalServer, "任务配置无效")
	}
	return config, nil
}

func validatePrompt(prompt string, maxLength int) error {
	if utf8.RuneCountInString(prompt) > maxLength {
		return errors.New(errors.ErrCodeBadRequest, fmt.Sprintf("提示词不能超过 %d 个字符", maxLength))
	}
	return nil
}

func (s *CreationService) prepareVideoReference(ctx context.Context, source string) (string, *objectstorage.StoredObject, error) {
	source = strings.TrimSpace(source)
	if source == "" {
		return "", nil, nil
	}
	if !strings.HasPrefix(source, "data:") {
		return source, nil, nil
	}
	stored, err := s.storageService.UploadBase64Image(ctx, source)
	if err != nil {
		return "", nil, err
	}
	url, err := s.storageService.TemporaryReferenceURL(ctx, stored)
	if err != nil {
		return "", nil, err
	}
	return url, stored, nil
}

func mediaAssetFromStored(userID int64, taskID *int64, stored *objectstorage.StoredObject) *model.MediaAsset {
	return &model.MediaAsset{TaskID: taskID, UserID: userID, StorageConfigID: stored.StorageConfigID, ResourceType: stored.ResourceType, ObjectKey: stored.ObjectKey, PublicURL: stored.PublicURL, ContentType: stored.ContentType, SizeBytes: stored.SizeBytes, ExpiresAt: stored.ExpiresAt, CreatedAt: time.Now()}
}

// checkConcurrentLimit 检查单用户进行中任务数限制。
func (s *CreationService) checkConcurrentLimit(userID int64, maxConcurrent int) error {
	count, err := s.taskRepo.CountProcessingTasks(userID)
	if err != nil {
		return err
	}

	if count >= int64(maxConcurrent) {
		return errors.ErrMaxConcurrentTasks
	}

	return nil
}

// deductCredit 扣除积分（事务中）
func (s *CreationService) deductCredit(tx *gorm.DB, userID int64, amount int, title string) error {
	// 1. 获取账户（加锁）
	account, err := s.creditRepo.GetAccountForUpdate(tx, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New(errors.ErrCodeInsufficientCredit, "积分账户不存在")
		}
		return err
	}

	// 2. 检查余额
	if account.Balance < amount {
		return errors.ErrInsufficientCredit
	}

	// 3. 扣除余额
	account.Balance -= amount
	account.TotalExpense += amount
	account.UpdatedAt = time.Now()
	if err := s.creditRepo.UpdateAccount(tx, account); err != nil {
		return err
	}

	// 4. 记录流水
	ledger := &model.CreditLedger{
		UserID:         userID,
		Type:           "expense",
		Amount:         amount,
		Title:          title,
		SourceType:     "task",
		BalanceAfter:   account.Balance,
		IdempotencyKey: fmt.Sprintf("task_expense_%d_%d", userID, time.Now().UnixNano()),
		CreatedAt:      time.Now(),
	}
	return s.creditRepo.CreateLedger(tx, ledger)
}

// taskToResponse 转换为响应
func (s *CreationService) taskToResponse(task *model.Task) *dto.TaskResponse {
	resp := &dto.TaskResponse{
		ID:           task.ID,
		Title:        task.Title,
		Type:         task.Type,
		Status:       task.Status,
		Progress:     task.Progress,
		Prompt:       task.Prompt,
		ModelName:    task.ModelName,
		ResultURL:    task.ResultURL,
		ThumbnailURL: task.ThumbnailURL,
		ErrorMsg:     task.ErrorMsg,
		Cost:         task.Cost,
		CreatedAt:    task.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if task.CompletedAt != nil {
		resp.CompletedAt = task.CompletedAt.Format("2006-01-02 15:04:05")
	}

	return resp
}
