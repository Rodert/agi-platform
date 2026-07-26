package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/worker/adapter"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	taskRepo         *repository.TaskRepository
	aiModelRepo      *repository.AIModelRepository
	channelModelRepo *repository.ChannelModelRepository
	assetRepo        *repository.MediaAssetRepository
	creditRepo       *repository.CreditRepository
	storageManager   *objectstorage.Manager
}

func NewImageProcessor(
	taskRepo *repository.TaskRepository,
	aiModelRepo *repository.AIModelRepository,
	channelModelRepo *repository.ChannelModelRepository,
	assetRepo *repository.MediaAssetRepository,
	creditRepo *repository.CreditRepository,
	storageManager *objectstorage.Manager,
) *ImageProcessor {
	return &ImageProcessor{
		taskRepo:         taskRepo,
		aiModelRepo:      aiModelRepo,
		channelModelRepo: channelModelRepo,
		assetRepo:        assetRepo,
		creditRepo:       creditRepo,
		storageManager:   storageManager,
	}
}

// Process 处理图片生成任务
func (p *ImageProcessor) Process(ctx context.Context, msg *TaskMessage) error {
	// 1. 获取任务
	task, err := p.taskRepo.FindByID(msg.TaskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %w", err)
	}

	// 2. A recovered upstream task must continue its original execution audit;
	// it is polling, not a newly generated retry.
	var attempt *model.TaskAttempt
	if task.ProviderTaskID != "" {
		attempt, err = p.taskRepo.FindOpenAttempt(task.ID)
	} else {
		attempt, err = p.taskRepo.StartAttempt(task)
	}
	if err != nil {
		return err
	}

	// 3. 获取 AI 模型配置
	aiModel, err := p.aiModelRepo.FindByName(msg.ModelName)
	if err != nil {
		return p.failTask(task, attempt, "AI 模型不存在")
	}

	// 4. Load the channel selected when the task was created. Legacy queued
	// tasks without a channel select the current highest-priority binding.
	channelID := task.ChannelID
	if channelID == 0 {
		binding, selectErr := p.channelModelRepo.SelectActiveChannel(aiModel.ID)
		if selectErr != nil {
			return p.failTask(task, attempt, "AI 模型当前没有可用渠道")
		}
		channelID = binding.ChannelID
	}
	channel, err := p.channelModelRepo.FindActiveChannel(channelID)
	if err != nil {
		return p.failTask(task, attempt, "任务渠道不可用")
	}

	// 5. 创建适配器
	adp, err := adapter.GetAdapter(aiModel, channel)
	if err != nil {
		return p.failTask(task, attempt, fmt.Sprintf("创建适配器失败: %v", err))
	}

	// 6. 更新进度
	task.Progress = 30
	p.taskRepo.Update(task)

	// 7. 调用 AI API
	logger.Info(fmt.Sprintf("🎨 调用 AI API: %s, Prompt: %s", aiModel.Name, msg.Prompt))

	request := &adapter.GenerateRequest{
		ModelName: aiModel.Name,
		Type:      task.Type,
		Prompt:    msg.Prompt,
		Params:    msg.Params,
	}
	var result *adapter.GenerateResponse
	if async, ok := adp.(adapter.AsyncTaskAdapter); ok {
		result, err = p.processAsyncTask(ctx, task, async, request)
	} else {
		result, err = adp.Generate(ctx, request)
	}
	if err != nil {
		return p.failTask(task, attempt, fmt.Sprintf("AI 生成失败: %v", err))
	}

	// 8. 更新进度
	task.Progress = 80
	p.taskRepo.Update(task)

	// 9. Store every generated result in the active object store.
	resourceType := "image"
	sourceURL := result.ImageURL
	if task.Type == "video" {
		resourceType, sourceURL = "video", result.VideoURL
	}
	if sourceURL == "" && result.ImageBase64 == "" {
		return p.failTask(task, attempt, "上游未返回可访问的生成结果")
	}
	task.Status, task.Progress, task.UpdatedAt = "uploading", 85, time.Now()
	if err := p.taskRepo.Update(task); err != nil {
		return p.failTask(task, attempt, "更新上传状态失败: "+err.Error())
	}
	var stored *objectstorage.StoredObject
	if result.ImageBase64 != "" && sourceURL == "" {
		stored, err = p.storageManager.UploadBase64(ctx, resourceType, result.ImageBase64)
	} else {
		stored, err = p.storageManager.UploadFromURL(ctx, resourceType, sourceURL)
	}
	if err != nil {
		return p.failTask(task, attempt, "上传对象存储失败: "+err.Error())
	}
	if stored.PublicURL == "" {
		return p.failTask(task, attempt, "生成资源未配置公开访问地址")
	}
	if err := p.assetRepo.Create(mediaAssetFromStored(task, stored)); err != nil {
		return p.failTask(task, attempt, "保存资源记录失败: "+err.Error())
	}
	resultURL := stored.PublicURL
	thumbnailURL := resultURL
	if task.Type == "video" && result.ThumbnailURL != "" && result.ThumbnailURL != sourceURL {
		thumbnail, uploadErr := p.storageManager.UploadFromURL(ctx, "thumbnail", result.ThumbnailURL)
		if uploadErr != nil {
			return p.failTask(task, attempt, "上传视频缩略图失败: "+uploadErr.Error())
		}
		if thumbnail.PublicURL == "" {
			return p.failTask(task, attempt, "视频缩略图未配置公开访问地址")
		}
		if uploadErr = p.assetRepo.Create(mediaAssetFromStored(task, thumbnail)); uploadErr != nil {
			return p.failTask(task, attempt, "保存视频缩略图记录失败: "+uploadErr.Error())
		}
		thumbnailURL = thumbnail.PublicURL
	}

	// 10. 更新任务为成功
	now := time.Now()
	task.Status = "success"
	task.Progress = 100
	task.ResultURL = resultURL
	task.ThumbnailURL = thumbnailURL
	task.UpdatedAt = now
	task.CompletedAt = &now

	if err := p.taskRepo.Update(task); err != nil {
		return err
	}
	if err := p.taskRepo.CompleteAttempt(attempt, "success", ""); err != nil {
		logger.Error("记录任务执行成功状态失败", zap.Error(err))
	}

	logger.Info(fmt.Sprintf("✅ 图片生成成功: TaskID=%d, URL=%s", task.ID, resultURL))

	// TODO: 发送 WebSocket 通知
	// TODO: 创建通知记录

	return nil
}

func mediaAssetFromStored(task *model.Task, stored *objectstorage.StoredObject) *model.MediaAsset {
	return &model.MediaAsset{TaskID: &task.ID, UserID: task.UserID, StorageConfigID: stored.StorageConfigID, ResourceType: stored.ResourceType, ObjectKey: stored.ObjectKey, PublicURL: stored.PublicURL, ContentType: stored.ContentType, SizeBytes: stored.SizeBytes, ExpiresAt: stored.ExpiresAt, CreatedAt: time.Now()}
}

// failTask 任务失败
func (p *ImageProcessor) failTask(task *model.Task, attempt *model.TaskAttempt, errorMsg string) error {
	logger.Error("任务失败", zap.String("error", errorMsg))

	now := time.Now()
	task.Status = "failed"
	task.ErrorMsg = errorMsg
	task.UpdatedAt = now
	task.CompletedAt = &now
	if err := p.taskRepo.Update(task); err != nil {
		return fmt.Errorf("更新任务失败状态失败: %w", err)
	}
	if attempt != nil {
		if err := p.taskRepo.CompleteAttempt(attempt, "failed", errorMsg); err != nil {
			logger.Error("记录任务执行失败状态失败", zap.Error(err))
		}
	}
	if task.AttemptCount > task.MaxRetryAttempts {
		if err := p.creditRepo.RefundFailedTask(task); err != nil {
			logger.Error("返还失败任务灵感值失败", zap.Int64("task_id", task.ID), zap.Error(err))
		}
	}

	return fmt.Errorf(errorMsg)
}

func (p *ImageProcessor) processAsyncTask(ctx context.Context, task *model.Task, adp adapter.AsyncTaskAdapter, request *adapter.GenerateRequest) (*adapter.GenerateResponse, error) {
	interval, timeout := 5*time.Second, 15*time.Minute
	if config, ok := adp.(adapter.PollingConfig); ok {
		interval, timeout = config.PollInterval(), config.PollTimeout()
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	pollCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if task.ProviderTaskID == "" {
		state, err := adp.Submit(pollCtx, request)
		if err != nil {
			return nil, err
		}
		if state.ProviderTaskID == "" {
			return nil, fmt.Errorf("上游未返回任务 ID")
		}
		if err := p.recordAsyncState(task, state, 35); err != nil {
			return nil, err
		}
		if state.Status == "failed" {
			return nil, asyncFailure(state)
		}
		if state.Status == "succeeded" {
			return asyncResult(state)
		}
	}

	for {
		select {
		case <-pollCtx.Done():
			return nil, fmt.Errorf("上游任务轮询超时: %w", pollCtx.Err())
		case <-time.After(interval):
		}
		state, err := adp.Poll(pollCtx, task.ProviderTaskID)
		if err != nil {
			return nil, fmt.Errorf("查询上游任务状态失败: %w", err)
		}
		if state.ProviderTaskID == "" {
			state.ProviderTaskID = task.ProviderTaskID
		}
		progress := state.Progress
		if progress <= 0 {
			progress = 60
		}
		if progress >= 100 && state.Status != "succeeded" {
			progress = 99
		}
		if err := p.recordAsyncState(task, state, progress); err != nil {
			return nil, err
		}
		switch state.Status {
		case "succeeded":
			return asyncResult(state)
		case "failed":
			return nil, asyncFailure(state)
		}
	}
}

func (p *ImageProcessor) recordAsyncState(task *model.Task, state *adapter.AsyncTask, fallbackProgress int) error {
	now := time.Now()
	task.ProviderTaskID = state.ProviderTaskID
	task.ProviderStatus = state.Status
	if len(state.RawResponse) > 0 {
		task.ProviderResponse = datatypes.JSON(append(json.RawMessage(nil), state.RawResponse...))
	}
	task.LastPolledAt = &now
	task.Status = "processing"
	task.Progress = fallbackProgress
	task.UpdatedAt = now
	return p.taskRepo.Update(task)
}

func asyncResult(state *adapter.AsyncTask) (*adapter.GenerateResponse, error) {
	if state.Result == nil || state.Result.VideoURL == "" {
		return nil, fmt.Errorf("上游任务已完成但未返回生成结果")
	}
	return state.Result, nil
}

func asyncFailure(state *adapter.AsyncTask) error {
	if state.ErrorMessage == "" {
		return fmt.Errorf("上游任务失败")
	}
	return fmt.Errorf("上游任务失败: %s", state.ErrorMessage)
}

func (p *ImageProcessor) MarkRetrying(taskID int64) error {
	return p.taskRepo.MarkRetryQueued(taskID)
}
