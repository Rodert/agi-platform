package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/dto"
	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/queue"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/errors"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CreationService struct {
	taskRepo       *repository.TaskRepository
	requestRepo    *repository.GenerationRequestRepository
	aiModelRepo    *repository.AIModelRepository
	creditRepo     *repository.CreditRepository
	configRepo     *repository.ConfigRepository
	storageService *StorageService
	queueProducer  *queue.Producer
	db             *gorm.DB
}

func NewCreationService(
	taskRepo *repository.TaskRepository,
	requestRepo *repository.GenerationRequestRepository,
	aiModelRepo *repository.AIModelRepository,
	creditRepo *repository.CreditRepository,
	configRepo *repository.ConfigRepository,
	storageService *StorageService,
	queueProducer *queue.Producer,
	db *gorm.DB,
) *CreationService {
	return &CreationService{
		taskRepo:       taskRepo,
		requestRepo:    requestRepo,
		aiModelRepo:    aiModelRepo,
		creditRepo:     creditRepo,
		configRepo:     configRepo,
		storageService: storageService,
		queueProducer:  queueProducer,
		db:             db,
	}
}

// CreateImageTask 创建图片生成任务
func (s *CreationService) CreateImageTask(userID int64, req *dto.CreateImageTaskRequest) (*dto.TaskResponse, error) {
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

	// 2. 校验模型参数并计算费用
	params, cost, err := s.resolveModelParams(aiModel, req.Params)
	if err != nil {
		return nil, err
	}

	// 3. 检查并发任务数限制
	maxConcurrent := 3 // 从配置读取
	if err := s.checkConcurrentLimit(userID, maxConcurrent); err != nil {
		return nil, err
	}

	// 4. 处理参考图（Base64 → URL）
	var referenceURL string
	if req.ReferenceImage != "" {
		url, err := s.storageService.UploadBase64Image(req.ReferenceImage)
		if err != nil {
			return nil, errors.NewWithDetails(errors.ErrCodeUploadFailed, "参考图上传失败", err.Error())
		}
		referenceURL = url
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
			UserID:    userID,
			RequestID: genReq.ID,
			Title:     utils.TruncateString(req.Prompt, 50),
			Type:      "image",
			Status:    "queued",
			Progress:  0,
			Prompt:    req.Prompt,
			ModelName: req.ModelName,
			Cost:      cost,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(task).Error; err != nil {
			return err
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

	// 2. 校验模型参数并计算费用
	params, cost, err := s.resolveModelParams(aiModel, req.Params)
	if err != nil {
		return nil, err
	}

	// 3. 检查并发任务数限制
	maxConcurrent := 3
	if err := s.checkConcurrentLimit(userID, maxConcurrent); err != nil {
		return nil, err
	}

	// 4. 构建参数
	if req.FirstFrameURL != "" {
		params["first_frame_url"] = req.FirstFrameURL
	}
	if req.LastFrameURL != "" {
		params["last_frame_url"] = req.LastFrameURL
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
			UserID:    userID,
			RequestID: genReq.ID,
			Title:     utils.TruncateString(req.Prompt, 50),
			Type:      "video",
			Status:    "queued",
			Progress:  0,
			Prompt:    req.Prompt,
			ModelName: req.ModelName,
			Cost:      cost,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(task).Error; err != nil {
			return err
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

// checkConcurrentLimit 检查并发任务数限制
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
