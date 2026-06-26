package service

import (
	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/config"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
)

type Dependencies struct {
	Config      config.Config
	Repos       repository.Repositories
	ProviderHub *provider.Registry
	Auth        auth.Manager
}

type Services struct {
	Auth         AuthService
	Admin        AdminService
	AdminCatalog AdminCatalogService
	APIKey       APIKeyService
	ImageModel   ImageModelService
	Provider     ProviderService
	ImageTask    ImageTaskService
}

func NewServices(deps Dependencies) Services {
	return Services{
		Auth:         NewAuthService(deps.Repos, deps.Auth, deps.Config.Auth.RegisterGiftCredits),
		Admin:        NewAdminService(deps.Repos, deps.Auth),
		AdminCatalog: NewAdminCatalogService(deps.Repos),
		APIKey:       NewAPIKeyService(deps.Repos, deps.Auth),
		ImageModel:   NewImageModelService(deps.Repos.ImageModels),
		Provider:     NewProviderService(deps.Repos.Providers),
		ImageTask:    NewImageTaskService(deps.Repos, deps.ProviderHub),
	}
}
