package router

import (
	"net/http"

	"agi-platform/server/internal/config"
	"agi-platform/server/internal/handler"
	"agi-platform/server/internal/middleware"

	"github.com/gin-gonic/gin"
)

func New(cfg config.Config, handlers handler.Handlers) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), middleware.CORS(), middleware.RequestID(), middleware.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.Auth.Register)
		api.POST("/auth/login", handlers.Auth.Login)
		api.GET("/models", handlers.ImageModel.List)
		api.GET("/video-models", handlers.Video.ListModels)
		api.GET("/assets/*key", handlers.Upload.Asset)
		api.HEAD("/assets/*key", handlers.Upload.Asset)

		userAPI := api.Group("")
		userAPI.Use(middleware.UserAuth(handlers.AuthManager))
		{
			userAPI.GET("/me", handlers.Auth.Me)
			userAPI.POST("/me/password", handlers.Auth.ChangePassword)
			userAPI.GET("/api-keys", handlers.APIKey.List)
			userAPI.POST("/api-keys", handlers.APIKey.Create)
			userAPI.DELETE("/api-keys/:id", handlers.APIKey.Revoke)
			userAPI.POST("/images/generate", handlers.ImageTask.Generate)
			userAPI.GET("/images/tasks", handlers.ImageTask.List)
			userAPI.GET("/images/tasks/:task_no", handlers.ImageTask.Get)
			userAPI.POST("/uploads/references", handlers.Upload.Reference)
			userAPI.GET("/wallet/logs", handlers.Wallet.List)
			userAPI.POST("/videos/generate", handlers.Video.Submit)
			userAPI.GET("/videos/tasks", handlers.Video.List)
			userAPI.GET("/videos/tasks/:task_no", handlers.Video.Get)
		}
	}

	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyAuth(handlers.Services.APIKey))
	{
		v1.POST("/images/generations", handlers.OpenAIImage.Generate)
	}

	admin := r.Group("/admin")
	{
		admin.POST("/auth/login", handlers.Admin.Login)

		adminAPI := admin.Group("")
		adminAPI.Use(middleware.AdminAuth(handlers.AuthManager))
		{
			adminAPI.GET("/me", handlers.Admin.Me)
			adminAPI.POST("/me/password", handlers.Admin.ChangePassword)
			adminAPI.GET("/users", handlers.Admin.ListUsers)
			adminAPI.POST("/users", handlers.Admin.CreateUser)
			adminAPI.PUT("/users/:id", handlers.Admin.UpdateUser)
			adminAPI.POST("/users/:id/credits", handlers.Admin.AdjustUserCredits)
			adminAPI.GET("/providers", handlers.AdminCatalog.ListProviders)
			adminAPI.POST("/providers", handlers.AdminCatalog.CreateProvider)
			adminAPI.PUT("/providers/:id", handlers.AdminCatalog.UpdateProvider)
			adminAPI.GET("/providers/:id/keys", handlers.AdminCatalog.ListProviderKeys)
			adminAPI.POST("/providers/:id/keys", handlers.AdminCatalog.CreateProviderKey)
			adminAPI.DELETE("/provider-keys/:id", handlers.AdminCatalog.DeleteProviderKey)
			adminAPI.GET("/image-models", handlers.AdminCatalog.ListImageModels)
			adminAPI.POST("/image-models", handlers.AdminCatalog.CreateImageModel)
			adminAPI.PUT("/image-models/:id", handlers.AdminCatalog.UpdateImageModel)
			adminAPI.DELETE("/image-models/:id", handlers.AdminCatalog.DeleteImageModel)
			adminAPI.GET("/image-models/:id/routes", handlers.AdminCatalog.ListImageModelRoutes)
			adminAPI.POST("/image-models/:id/routes", handlers.AdminCatalog.CreateImageModelRoute)
			adminAPI.PUT("/image-model-routes/:id", handlers.AdminCatalog.UpdateImageModelRoute)
			adminAPI.GET("/video-models", handlers.AdminCatalog.ListVideoModels)
			adminAPI.POST("/video-models", handlers.AdminCatalog.CreateVideoModel)
			adminAPI.PUT("/video-models/:id", handlers.AdminCatalog.UpdateVideoModel)
			adminAPI.DELETE("/video-models/:id", handlers.AdminCatalog.DeleteVideoModel)
			adminAPI.GET("/video-models/:id/routes", handlers.AdminCatalog.ListVideoModelRoutes)
			adminAPI.POST("/video-models/:id/routes", handlers.AdminCatalog.CreateVideoModelRoute)
			adminAPI.PUT("/video-model-routes/:id", handlers.AdminCatalog.UpdateVideoModelRoute)
			adminAPI.POST("/upstream/models/query", handlers.AdminCatalog.QueryUpstreamModels)
			adminAPI.GET("/image-tasks", handlers.ImageTask.AdminList)
			adminAPI.GET("/image-tasks/:task_no", handlers.ImageTask.AdminGet)
			adminAPI.GET("/video-tasks", handlers.Video.AdminList)
			adminAPI.GET("/video-tasks/:task_no", handlers.Video.AdminGet)
			adminAPI.GET("/wallet/logs", handlers.Wallet.AdminList)
			adminAPI.GET("/database/tables", handlers.Database.ListTables)
			adminAPI.GET("/database/tables/:table", handlers.Database.GetTable)
		}
	}

	return r
}
