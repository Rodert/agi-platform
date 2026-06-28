package handler

import (
	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/service"
	"agi-platform/server/internal/storage"
)

type Handlers struct {
	Auth         AuthHandler
	Admin        AdminAuthHandler
	AdminCatalog AdminCatalogHandler
	APIKey       APIKeyHandler
	ImageModel   ImageModelHandler
	Provider     ProviderHandler
	ImageTask    ImageTaskHandler
	Upload       UploadHandler
	Wallet       WalletHandler
	Video        VideoHandler
	Database     DatabaseHandler
	OpenAIImage  OpenAIImageHandler
	AuthManager  auth.Manager
	Services     service.Services
}

func NewHandlers(services service.Services, authManager auth.Manager, objectStore storage.Store) Handlers {
	imageTask := NewImageTaskHandler(services.ImageTask)
	return Handlers{
		Auth:         NewAuthHandler(services.Auth),
		Admin:        NewAdminAuthHandler(services.Admin),
		AdminCatalog: NewAdminCatalogHandler(services.AdminCatalog),
		APIKey:       NewAPIKeyHandler(services.APIKey),
		ImageModel:   NewImageModelHandler(services.ImageModel),
		Provider:     NewProviderHandler(services.Provider),
		ImageTask:    imageTask,
		Upload:       NewUploadHandler(objectStore),
		Wallet:       NewWalletHandler(services.Wallet),
		Video:        NewVideoHandler(services.Video),
		Database:     NewDatabaseHandler(services.Database),
		OpenAIImage:  NewOpenAIImageHandler(services.ImageTask),
		AuthManager:  authManager,
		Services:     services,
	}
}
