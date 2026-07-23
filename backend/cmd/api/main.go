package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/javapub/agi-platform-backend/internal/handler"
	"github.com/javapub/agi-platform-backend/internal/middleware"
	"github.com/javapub/agi-platform-backend/internal/objectstorage"
	"github.com/javapub/agi-platform-backend/internal/queue"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/internal/service"
	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/database"
	"github.com/javapub/agi-platform-backend/pkg/logger"
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

	logger.Info("🚀 启动 AGI Platform API 服务...")

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

	// 提示数据库迁移
	if cfg.Server.Env == "development" {
		logger.Info("💡 请确保已执行数据库迁移脚本")
		logger.Info("   mysql -u root -p agi_platform < scripts/migrations/001_create_tables.sql")
		logger.Info("   mysql -u root -p agi_platform < scripts/seeds/seed.sql")
	}

	// 设置 Gin 模式
	if !cfg.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := setupRouter(cfg)

	// 创建 HTTP 服务器
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// 启动服务器（协程）
	go func() {
		logger.Info(fmt.Sprintf("✅ API 服务启动成功，监听端口: %d", cfg.Server.Port))
		logger.Info(fmt.Sprintf("📖 环境: %s", cfg.Server.Env))
		logger.Info(fmt.Sprintf("🌐 访问地址: http://localhost:%d", cfg.Server.Port))

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(fmt.Sprintf("服务器启动失败: %v", err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 正在关闭服务器...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("服务器强制关闭: %v", err))
	}

	logger.Info("✅ 服务器已关闭")
}

// setupRouter 配置路由
func setupRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()

	// 全局中间件
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware(&cfg.System))

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "agi-platform-api",
			"time":    time.Now().Unix(),
		})
	})
	router.Static("/uploads", "./uploads")

	// 初始化依赖并注册路由
	initHandlers(cfg, router)

	return router
}

// initHandlers 初始化所有 Handler
func initHandlers(cfg *config.Config, router *gin.Engine) {
	// Repository
	userRepo := repository.NewUserRepository(database.DB)
	codeRepo := repository.NewVerificationCodeRepository(database.DB)
	configRepo := repository.NewConfigRepository(database.DB)
	taskRepo := repository.NewTaskRepository(database.DB)
	requestRepo := repository.NewGenerationRequestRepository(database.DB)
	aiModelRepo := repository.NewAIModelRepository(database.DB)
	providerAccountRepo := repository.NewAIProviderAccountRepository(database.DB)
	channelModelRepo := repository.NewChannelModelRepository(database.DB)
	creditRepo := repository.NewCreditRepository(database.DB)

	adminRepo := repository.NewAdminRepository(database.DB)
	workRepo := repository.NewWorkRepository(database.DB)
	storageConfigRepo := repository.NewStorageConfigRepository(database.DB)
	resourcePolicyRepo := repository.NewResourcePolicyRepository(database.DB)
	mediaAssetRepo := repository.NewMediaAssetRepository(database.DB)
	announcementRepo := repository.NewAnnouncementRepository(database.DB)
	promptOptimizationRepo := repository.NewPromptOptimizationRepository(database.DB)

	// Queue
	queueProducer := queue.NewProducer(database.RDB, cfg.Worker.RedisStream)

	// Service
	creditService := service.NewCreditService(database.DB)
	inviteService := service.NewInvitationService(database.DB)
	emailService := service.NewEmailService(configRepo)
	objectStorageManager := objectstorage.NewManager(storageConfigRepo, resourcePolicyRepo)
	storageService := service.NewStorageService(objectStorageManager)
	authService := service.NewAuthService(userRepo, codeRepo, configRepo, creditService, inviteService, emailService, &cfg.JWT, database.DB)
	userService := service.NewUserService(userRepo, creditRepo)
	creationService := service.NewCreationService(taskRepo, requestRepo, aiModelRepo, channelModelRepo, creditRepo, mediaAssetRepo, configRepo, storageService, queueProducer, database.DB)
	adminService := service.NewAdminService(adminRepo, workRepo, taskRepo, userRepo, creditRepo, mediaAssetRepo, objectStorageManager, &cfg.JWT, database.DB)
	workService := service.NewWorkService(workRepo, taskRepo, userRepo)
	storageConfigService := service.NewStorageConfigService(storageConfigRepo)
	resourcePolicyService := service.NewResourcePolicyService(resourcePolicyRepo)
	channelCatalogService := service.NewChannelCatalogService(providerAccountRepo, aiModelRepo, channelModelRepo)
	announcementService := service.NewAnnouncementService(announcementRepo)
	promptOptimizationService := service.NewPromptOptimizationService(configRepo, aiModelRepo, channelModelRepo, creditRepo, promptOptimizationRepo, database.DB)

	// Handler
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	creationHandler := handler.NewCreationHandler(creationService)
	adminHandler := handler.NewAdminHandler(adminService)
	workHandler := handler.NewWorkHandler(workService)
	adminConfigHandler := handler.NewAdminConfigHandler(configRepo, aiModelRepo, providerAccountRepo, channelModelRepo, channelCatalogService, storageConfigService, resourcePolicyService, emailService)
	announcementHandler := handler.NewAnnouncementHandler(announcementService)
	promptOptimizationHandler := handler.NewPromptOptimizationHandler(promptOptimizationService)

	// 注册路由
	registerRoutes(router, cfg, authHandler, userHandler, creationHandler, adminHandler, workHandler, adminConfigHandler, announcementHandler, promptOptimizationHandler)
}

// registerRoutes 注册路由
func registerRoutes(router *gin.Engine, cfg *config.Config, authHandler *handler.AuthHandler, userHandler *handler.UserHandler, creationHandler *handler.CreationHandler, adminHandler *handler.AdminHandler, workHandler *handler.WorkHandler, adminConfigHandler *handler.AdminConfigHandler, announcementHandler *handler.AnnouncementHandler, promptOptimizationHandler *handler.PromptOptimizationHandler) {
	// API 路由组
	apiV1 := router.Group("/api/v1")
	{
		// 公开接口
		apiV1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "pong"})
		})

		// 认证接口（不需要登录）
		auth := apiV1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/send-code", authHandler.SendCode)
		}

		// 用户接口（需要登录）
		users := apiV1.Group("/users")
		users.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PATCH("/profile", userHandler.UpdateProfile)
		}

		// 模型列表公开，创建任务需要登录。
		apiV1.GET("/generation/models", creationHandler.GetModels)
		apiV1.GET("/announcements", announcementHandler.ListPublished)

		// 创作接口（需要登录）
		generation := apiV1.Group("/generation")
		generation.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			generation.POST("/image", creationHandler.CreateImageTask)
			generation.POST("/video", creationHandler.CreateVideoTask)
			generation.POST("/prompt-optimization", promptOptimizationHandler.Optimize)
		}

		// 任务接口（需要登录）
		tasks := apiV1.Group("/tasks")
		tasks.Use(middleware.AuthMiddleware(&cfg.JWT))
		{
			tasks.GET("", creationHandler.GetTaskList)
			tasks.GET("/:id", creationHandler.GetTask)
			tasks.GET("/:id/download", creationHandler.DownloadTask)
		}

		// 作品接口
		works := apiV1.Group("/works")
		{
			works.GET("", workHandler.GetWorkList) // 获取作品列表（公开）
			works.GET("/:id", workHandler.GetWork) // 获取作品详情（公开）

			// 需要登录的接口
			worksAuth := works.Group("")
			worksAuth.Use(middleware.AuthMiddleware(&cfg.JWT))
			{
				// 发布作品
				worksAuth.POST("", workHandler.PublishWork)

				// 点赞
				worksAuth.POST("/:id/like", workHandler.LikeWork)
				worksAuth.DELETE("/:id/like", workHandler.UnlikeWork)

				// 收藏（同时支持 collect 和 favorite 两个路径，保持兼容）
				worksAuth.POST("/:id/collect", workHandler.CollectWork)
				worksAuth.DELETE("/:id/collect", workHandler.UncollectWork)
				worksAuth.POST("/:id/favorite", workHandler.CollectWork)     // 前端兼容
				worksAuth.DELETE("/:id/favorite", workHandler.UncollectWork) // 前端兼容
			}
		}

		// TODO: 积分模块路由
		// credits := apiV1.Group("/credits")
	}

	// 管理后台路由
	adminV1 := router.Group("/admin/v1")
	{
		adminV1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "admin pong"})
		})

		// 管理员认证（不需要登录）
		auth := adminV1.Group("/auth")
		{
			auth.POST("/login", adminHandler.Login)
		}

		// 需要管理员登录的接口
		adminV1.Use(middleware.AdminAuthMiddleware(&cfg.JWT))
		{
			adminV1.GET("/stats", adminHandler.GetStats)
			config := adminV1.Group("/config")
			{
				config.GET("/basic", adminConfigHandler.GetBasic)
				config.PUT("/basic", adminConfigHandler.SaveBasic)
				config.GET("/email", adminConfigHandler.GetEmail)
				config.PUT("/email", adminConfigHandler.SaveEmail)
				config.POST("/email/test", adminConfigHandler.TestEmail)
				config.GET("/task", adminConfigHandler.GetTaskConfig)
				config.PUT("/task", adminConfigHandler.SaveTaskConfig)
				config.GET("/prompt-optimization", adminConfigHandler.GetPromptOptimizationConfig)
				config.PUT("/prompt-optimization", adminConfigHandler.SavePromptOptimizationConfig)
				config.GET("/prompt-optimization/logs", promptOptimizationHandler.ListAdmin)
				config.GET("/models", adminConfigHandler.GetModels)
				config.PUT("/models/:id", adminConfigHandler.UpdateModel)
				config.PUT("/models/:id/status", adminConfigHandler.UpdateModelStatus)
				config.GET("/storage", adminConfigHandler.GetStorage)
				config.POST("/storage", adminConfigHandler.CreateStorage)
				config.PUT("/storage/:id", adminConfigHandler.UpdateStorage)
				config.POST("/storage/:id/enable", adminConfigHandler.EnableStorage)
				config.DELETE("/storage/:id", adminConfigHandler.DeleteStorage)
				config.GET("/storage/policies", adminConfigHandler.GetResourcePolicies)
				config.PUT("/storage/policies/:type", adminConfigHandler.UpdateResourcePolicy)
			}
			accounts := adminV1.Group("/provider-accounts")
			{
				accounts.GET("", adminConfigHandler.ListProviderAccounts)
				accounts.POST("", adminConfigHandler.CreateProviderAccount)
				accounts.PUT("/:id", adminConfigHandler.UpdateProviderAccount)
				accounts.DELETE("/:id", adminConfigHandler.DeleteProviderAccount)
			}
			channels := adminV1.Group("/channels")
			{
				channels.GET("", adminConfigHandler.ListChannels)
				channels.POST("", adminConfigHandler.CreateChannel)
				channels.PUT("/:id", adminConfigHandler.UpdateChannel)
				channels.DELETE("/:id", adminConfigHandler.DeleteChannel)
				channels.POST("/:id/sync-models", adminConfigHandler.SyncChannelModels)
				channels.POST("/:id/models", adminConfigHandler.BindChannelModel)
				channels.PUT("/:id/models/:modelID/status", adminConfigHandler.UpdateChannelModelStatus)
			}

			// 用户管理
			users := adminV1.Group("/users")
			{
				users.GET("", adminHandler.GetUserList)
				users.POST("", adminHandler.CreateUser)
				users.PUT("/:id", adminHandler.UpdateUser)
				users.POST("/:id/credits", middleware.RequireRole("admin"), adminHandler.RechargeUserCredit)
				users.PUT("/:id/status", adminHandler.UpdateUserStatus)
			}
			tasks := adminV1.Group("/tasks")
			{
				tasks.GET("", adminHandler.GetTaskList)
			}
			announcements := adminV1.Group("/announcements")
			{
				announcements.GET("", announcementHandler.ListAdmin)
				announcements.POST("", announcementHandler.Create)
				announcements.PUT("/:id", announcementHandler.Update)
				announcements.DELETE("/:id", announcementHandler.Delete)
			}
		}

		// 作品审核
		works := adminV1.Group("/works")
		works.Use(middleware.AdminAuthMiddleware(&cfg.JWT))
		{
			works.GET("/pending", adminHandler.GetPendingWorks)
			works.POST("/:id/audit", adminHandler.AuditWork)
		}
	}
}
