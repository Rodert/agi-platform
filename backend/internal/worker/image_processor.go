package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/worker/adapter"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"go.uber.org/zap"
)

// ImageProcessor 图片处理器
type ImageProcessor struct {
	taskRepo    *repository.TaskRepository
	aiModelRepo *repository.AIModelRepository
}

func NewImageProcessor(
	taskRepo *repository.TaskRepository,
	aiModelRepo *repository.AIModelRepository,
) *ImageProcessor {
	return &ImageProcessor{
		taskRepo:    taskRepo,
		aiModelRepo: aiModelRepo,
	}
}

// Process 处理图片生成任务
func (p *ImageProcessor) Process(ctx context.Context, msg *TaskMessage) error {
	// 1. 获取任务
	task, err := p.taskRepo.FindByID(msg.TaskID)
	if err != nil {
		return fmt.Errorf("任务不存在: %w", err)
	}

	// 2. 更新状态为处理中
	task.Status = "processing"
	task.Progress = 10
	task.UpdatedAt = time.Now()
	if err := p.taskRepo.Update(task); err != nil {
		return err
	}

	// 3. 获取 AI 模型配置
	aiModel, err := p.aiModelRepo.FindByName(msg.ModelName)
	if err != nil {
		return p.failTask(task, "AI 模型不存在")
	}

	// 4. 创建适配器
	adp, err := adapter.GetAdapter(aiModel)
	if err != nil {
		return p.failTask(task, fmt.Sprintf("创建适配器失败: %v", err))
	}

	// 5. 更新进度
	task.Progress = 30
	p.taskRepo.Update(task)

	// 6. 调用 AI API
	logger.Info(fmt.Sprintf("🎨 调用 AI API: %s, Prompt: %s", aiModel.Name, msg.Prompt))

	result, err := adp.Generate(ctx, &adapter.GenerateRequest{
		Prompt: msg.Prompt,
		Params: msg.Params,
	})
	if err != nil {
		return p.failTask(task, fmt.Sprintf("AI 生成失败: %v", err))
	}

	// 7. 更新进度
	task.Progress = 80
	p.taskRepo.Update(task)

	// 8. 处理结果（Base64 或 URL）
	resultURL := result.ImageURL
	if resultURL == "" && result.ImageBase64 != "" {
		// TODO: 上传 Base64 到对象存储
		// resultURL = uploadBase64ToStorage(result.ImageBase64)
		resultURL = "https://cdn.tide.ai/temp/demo-image.jpg" // 临时
	}

	// 9. 更新任务为成功
	now := time.Now()
	task.Status = "success"
	task.Progress = 100
	task.ResultURL = resultURL
	task.ThumbnailURL = resultURL // 临时使用相同URL
	task.UpdatedAt = now
	task.CompletedAt = &now

	if err := p.taskRepo.Update(task); err != nil {
		return err
	}

	logger.Info(fmt.Sprintf("✅ 图片生成成功: TaskID=%d, URL=%s", task.ID, resultURL))

	// TODO: 发送 WebSocket 通知
	// TODO: 创建通知记录

	return nil
}

// failTask 任务失败
func (p *ImageProcessor) failTask(task *model.Task, errorMsg string) error {
	logger.Error("任务失败", zap.String("error", errorMsg))

	now := time.Now()
	task.Status = "failed"
	task.ErrorMsg = errorMsg
	task.UpdatedAt = now
	task.CompletedAt = &now
	p.taskRepo.Update(task)

	return fmt.Errorf(errorMsg)
}
