package app

import (
	"fmt"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/config"
	"agi-platform/server/internal/database"
	"agi-platform/server/internal/handler"
	"agi-platform/server/internal/provider"
	"agi-platform/server/internal/repository"
	"agi-platform/server/internal/router"
	"agi-platform/server/internal/service"
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
	authManager := auth.NewManager(auth.Config{
		JWTSecret:     cfg.Auth.JWTSecret,
		TokenLifetime: cfg.Auth.TokenLifetime,
	})

	services := service.NewServices(service.Dependencies{
		Config:      cfg,
		Repos:       repos,
		ProviderHub: providers,
		Auth:        authManager,
	})
	handlers := handler.NewHandlers(services, authManager)

	engine := router.New(cfg, handlers)
	return engine.Run(cfg.HTTP.Addr())
}
