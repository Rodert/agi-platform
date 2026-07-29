package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/queue"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/worker"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/database"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志
	logger.Init(cfg.Server.Debug)
	defer logger.Sync()

	logger.Info("🚀 启动 AGI Platform Worker 服务...")

	// 初始化数据库
	_, err = database.InitMySQL(&cfg.Database)
	if err != nil {
		logger.Fatal(fmt.Sprintf("初始化数据库失败: %v", err))
	}
	defer database.Close()

	// 初始化 Redis
	_, err = database.InitRedis(&cfg.Redis)
	if err != nil {
		logger.Fatal(fmt.Sprintf("初始化 Redis 失败: %v", err))
	}
	defer database.CloseRedis()

	// 创建 Repository
	taskRepo := repository.NewTaskRepository(database.DB)
	aiModelRepo := repository.NewAIModelRepository(database.DB)
	channelModelRepo := repository.NewChannelModelRepository(database.DB)
	storageConfigRepo := repository.NewStorageConfigRepository(database.DB)
	resourcePolicyRepo := repository.NewResourcePolicyRepository(database.DB)
	mediaAssetRepo := repository.NewMediaAssetRepository(database.DB)
	creditRepo := repository.NewCreditRepository(database.DB)
	configRepo := repository.NewConfigRepository(database.DB)
	objectStorageManager := objectstorage.NewManager(storageConfigRepo, resourcePolicyRepo, cfg.Server.PublicBaseURL)

	// 创建 Processor
	imageProcessor := worker.NewImageProcessor(taskRepo, aiModelRepo, channelModelRepo, mediaAssetRepo, creditRepo, objectStorageManager)
	assetCleaner := worker.NewAssetCleaner(mediaAssetRepo, objectStorageManager)
	queueProducer := queue.NewProducer(database.RDB, cfg.Worker.ImageRedisStream, cfg.Worker.VideoRedisStream, cfg.Worker.RedisStream)
	const interruptedSubmissionError = "上游提交结果未知：服务重启或请求中断，未收到上游响应"

	// A non-idempotent upstream must never receive an automatic second submit.
	// Fail and refund tasks that crashed after submission began but before a
	// provider task ID (or synchronous result) could be persisted.
	if tasks, recoverErr := taskRepo.FindInterruptedSubmissions(); recoverErr != nil {
		logger.Error("查询提交结果未知任务失败", zap.Error(recoverErr))
	} else {
		for _, task := range tasks {
			updated, failErr := taskRepo.FailInterruptedSubmission(task, interruptedSubmissionError)
			if failErr != nil {
				logger.Error("标记提交结果未知任务失败", zap.Int64("task_id", task.ID), zap.Error(failErr))
				continue
			}
			if updated {
				if refundErr := creditRepo.RefundFailedTask(task); refundErr != nil {
					logger.Error("返还提交结果未知任务灵感值失败", zap.Int64("task_id", task.ID), zap.Error(refundErr))
				}
				logger.Warn("任务上游提交结果未知，已失败并退款", zap.Int64("task_id", task.ID))
			}
		}
	}

	// Redis stream entries that were in-flight during a restart are not claimed
	// by this consumer. Persisted provider IDs let us safely enqueue polling
	// again without creating a second upstream generation.
	if tasks, recoverErr := taskRepo.FindPendingProviderTasks(); recoverErr != nil {
		logger.Error("查询待恢复上游任务失败", zap.Error(recoverErr))
	} else {
		for _, task := range tasks {
			params := map[string]interface{}{}
			if task.Request != nil && len(task.Request.Params) > 0 {
				if err := json.Unmarshal(task.Request.Params, &params); err != nil {
					logger.Error("解析待恢复任务参数失败", zap.Int64("task_id", task.ID), zap.Error(err))
					continue
				}
			}
			if err := queueProducer.Publish(context.Background(), &queue.TaskMessage{TaskID: task.ID, UserID: task.UserID, Type: task.Type, Prompt: task.Prompt, ModelName: task.ModelName, Params: params}); err != nil {
				logger.Error("重新投递上游任务失败", zap.Int64("task_id", task.ID), zap.Error(err))
			}
		}
	}
	if tasks, recoverErr := taskRepo.FindQueuedTasks(); recoverErr != nil {
		logger.Error("查询待恢复排队任务失败", zap.Error(recoverErr))
	} else {
		for _, task := range tasks {
			params := map[string]interface{}{}
			if task.Request != nil && len(task.Request.Params) > 0 {
				if err := json.Unmarshal(task.Request.Params, &params); err != nil {
					logger.Error("解析待恢复排队任务参数失败", zap.Int64("task_id", task.ID), zap.Error(err))
					continue
				}
			}
			if err := queueProducer.Publish(context.Background(), &queue.TaskMessage{TaskID: task.ID, UserID: task.UserID, Type: task.Type, Prompt: task.Prompt, ModelName: task.ModelName, Params: params}); err != nil {
				logger.Error("重新投递排队任务失败", zap.Int64("task_id", task.ID), zap.Error(err))
			}
		}
	}

	imageStream, videoStream := cfg.Worker.ImageRedisStream, cfg.Worker.VideoRedisStream
	if imageStream == "" {
		imageStream = cfg.Worker.RedisStream
	}
	if videoStream == "" {
		videoStream = cfg.Worker.RedisStream
	}
	imageConcurrency, videoConcurrency := cfg.Worker.ImageConcurrency, cfg.Worker.VideoConcurrency
	if taskConfig, configErr := configRepo.GetTaskConfig(); configErr != nil {
		logger.Error("读取任务并发配置失败，使用文件配置", zap.Error(configErr))
	} else {
		if taskConfig.ImageConcurrency > 0 {
			imageConcurrency = taskConfig.ImageConcurrency
		}
		if taskConfig.VideoConcurrency > 0 {
			videoConcurrency = taskConfig.VideoConcurrency
		}
	}
	if imageConcurrency < 1 {
		imageConcurrency = cfg.Worker.Concurrency
	}
	if imageConcurrency < 1 {
		imageConcurrency = 1
	}
	if videoConcurrency < 1 {
		videoConcurrency = 1
	}
	imageConsumer := queue.NewConsumer(database.RDB, imageStream, "agi-image-workers", fmt.Sprintf("image-worker-%d", os.Getpid()), imageConcurrency, imageProcessor)
	videoConsumer := queue.NewConsumer(database.RDB, videoStream, "agi-video-workers", fmt.Sprintf("video-worker-%d", os.Getpid()), videoConcurrency, imageProcessor)

	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runExpiredAssetCleanup(ctx, assetCleaner)
	if err := imageConsumer.Start(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("启动图片消费者失败: %v", err))
	}
	if videoStream != imageStream {
		if err := videoConsumer.Start(ctx); err != nil {
			logger.Fatal(fmt.Sprintf("启动视频消费者失败: %v", err))
		}
	}

	logger.Info(fmt.Sprintf("✅ Worker 服务启动成功，图片并发: %d，视频并发: %d", imageConcurrency, videoConcurrency))
	logger.Info(fmt.Sprintf("📨 监听队列: 图片=%s，视频=%s", imageStream, videoStream))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 正在关闭 Worker 服务...")

	imageConsumer.Stop()
	if videoStream != imageStream {
		videoConsumer.Stop()
	}
	cancel()

	time.Sleep(1 * time.Second) // 等待处理中的任务完成

	logger.Info("✅ Worker 服务已关闭")
}

func runExpiredAssetCleanup(ctx context.Context, cleaner *worker.AssetCleaner) {
	cleanup := func() {
		if err := cleaner.Run(ctx); err != nil {
			logger.Error("清理过期生成资源失败", zap.Error(err))
		}
	}
	cleanup()
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cleanup()
		}
	}
}
