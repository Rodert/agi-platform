package app

import (
	"context"
	"fmt"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/config"
	"agi-platform/server/internal/database"
	"agi-platform/server/internal/handler"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
	"agi-platform/server/internal/router"
	"agi-platform/server/internal/service"
	"agi-platform/server/internal/storage"
)

func Run() error {
	cfg := config.Load()

	db, err := database.Open(cfg.Database)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	repos := repository.NewRepositories(db)
	providers := provider.NewRegistry()
	providers.Register(provider.NewMockProvider())
	providers.Register(provider.NewOpenAICompatibleProvider("openai-compatible"))
	providers.Register(provider.NewOpenAICompatibleProvider("openai"))
	authManager := auth.NewManager(auth.Config{
		JWTSecret:     cfg.Auth.JWTSecret,
		TokenLifetime: cfg.Auth.TokenLifetime,
	})
	if err := service.EnsureBootstrapAdmin(context.Background(), repos, authManager, cfg.Admin); err != nil {
		return fmt.Errorf("ensure bootstrap admin: %w", err)
	}
	objectStore, err := storage.NewFromConfig(cfg.Storage)
	if err != nil {
		return fmt.Errorf("init storage: %w", err)
	}

	services := service.NewServices(service.Dependencies{
		Config:      cfg,
		Repos:       repos,
		ProviderHub: providers,
		Auth:        authManager,
		Storage:     objectStore,
	})
	handlers := handler.NewHandlers(services, authManager, objectStore)

	engine := router.New(cfg, handlers)
	return engine.Run(cfg.HTTP.Addr())
}
