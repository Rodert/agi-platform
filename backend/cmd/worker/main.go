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
	objectStorageManager := objectstorage.NewManager(storageConfigRepo, resourcePolicyRepo, cfg.Server.PublicBaseURL)

	// 创建 Processor
	imageProcessor := worker.NewImageProcessor(taskRepo, aiModelRepo, channelModelRepo, mediaAssetRepo, creditRepo, objectStorageManager)
	assetCleaner := worker.NewAssetCleaner(mediaAssetRepo, objectStorageManager)
	queueProducer := queue.NewProducer(database.RDB, cfg.Worker.RedisStream)

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

	// 创建 Consumer
	consumer := queue.NewConsumer(
		database.RDB,
		cfg.Worker.RedisStream,
		"agi-workers",
		fmt.Sprintf("worker-%d", os.Getpid()),
		imageProcessor,
	)

	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go runExpiredAssetCleanup(ctx, assetCleaner)
	if err := consumer.Start(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("启动消费者失败: %v", err))
	}

	logger.Info(fmt.Sprintf("✅ Worker 服务启动成功，并发数: %d", cfg.Worker.Concurrency))
	logger.Info(fmt.Sprintf("📨 监听队列: %s", cfg.Worker.RedisStream))

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 正在关闭 Worker 服务...")

	consumer.Stop()
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
