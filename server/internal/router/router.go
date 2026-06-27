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

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.Auth.Register)
		api.POST("/auth/login", handlers.Auth.Login)
		api.GET("/models", handlers.ImageModel.List)

		userAPI := api.Group("")
		userAPI.Use(middleware.UserAuth(handlers.AuthManager))
		{
			userAPI.GET("/me", handlers.Auth.Me)
			userAPI.GET("/api-keys", handlers.APIKey.List)
			userAPI.POST("/api-keys", handlers.APIKey.Create)
			userAPI.DELETE("/api-keys/:id", handlers.APIKey.Revoke)
			userAPI.POST("/images/generate", handlers.ImageTask.Generate)
			userAPI.GET("/images/tasks/:task_no", handlers.ImageTask.Get)
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
			adminAPI.GET("/users", handlers.Admin.ListUsers)
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
			adminAPI.GET("/image-models/:id/routes", handlers.AdminCatalog.ListImageModelRoutes)
			adminAPI.POST("/image-models/:id/routes", handlers.AdminCatalog.CreateImageModelRoute)
			adminAPI.PUT("/image-model-routes/:id", handlers.AdminCatalog.UpdateImageModelRoute)
			adminAPI.GET("/image-tasks", handlers.ImageTask.AdminList)
		}
	}

	return r
}
