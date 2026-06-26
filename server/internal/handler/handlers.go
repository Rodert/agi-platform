package handler

import (
	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/service"
)

type Handlers struct {
	Auth         AuthHandler
	Admin        AdminAuthHandler
	AdminCatalog AdminCatalogHandler
	APIKey       APIKeyHandler
	ImageModel   ImageModelHandler
	Provider     ProviderHandler
	ImageTask    ImageTaskHandler
	OpenAIImage  OpenAIImageHandler
	AuthManager  auth.Manager
	Services     service.Services
}

func NewHandlers(services service.Services, authManager auth.Manager) Handlers {
	imageTask := NewImageTaskHandler(services.ImageTask)
	return Handlers{
		Auth:         NewAuthHandler(services.Auth),
		Admin:        NewAdminAuthHandler(services.Admin),
		AdminCatalog: NewAdminCatalogHandler(services.AdminCatalog),
		APIKey:       NewAPIKeyHandler(services.APIKey),
		ImageModel:   NewImageModelHandler(services.ImageModel),
		Provider:     NewProviderHandler(services.Provider),
		ImageTask:    imageTask,
		OpenAIImage:  NewOpenAIImageHandler(services.ImageTask),
		AuthManager:  authManager,
		Services:     services,
	}
}
