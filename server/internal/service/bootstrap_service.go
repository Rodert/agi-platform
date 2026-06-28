package service

import (
	"context"
	"errors"
	"strings"

	"agi-platform/server/internal/auth"
	"agi-platform/server/internal/config"
	"agi-platform/server/internal/model"
	"agi-platform/server/internal/repository"
)

func EnsureBootstrapAdmin(ctx context.Context, repos repository.Repositories, authManager auth.Manager, cfg config.AdminConfig) error {
	if !cfg.Enabled {
		return nil
	}
	username := strings.TrimSpace(cfg.Username)
	if username == "" {
		return ErrInvalidRequest
	}

	admin, err := repos.Admins.FindByUsername(ctx, username)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		return err
	}

	nickname := strings.TrimSpace(cfg.Nickname)
	if nickname == "" {
		nickname = username
	}
	role := strings.TrimSpace(cfg.Role)
	if role == "" {
		role = "super_admin"
	}
	status := normalizeUserStatus(cfg.Status)

	if errors.Is(err, repository.ErrNotFound) {
		if len(cfg.Password) < 6 {
			return ErrInvalidRequest
		}
		passwordHash, err := authManager.HashPassword(cfg.Password)
		if err != nil {
			return err
		}
		return repos.Admins.Create(ctx, &model.AdminUser{
			Username:     username,
			PasswordHash: passwordHash,
			Nickname:     nickname,
			Role:         role,
			Status:       status,
		})
	}

	admin.Nickname = nickname
	admin.Role = role
	admin.Status = status
	if cfg.ResetPasswordOnStartup && strings.TrimSpace(cfg.Password) != "" {
		if len(cfg.Password) < 6 {
			return ErrInvalidRequest
		}
		passwordHash, err := authManager.HashPassword(cfg.Password)
		if err != nil {
			return err
		}
		admin.PasswordHash = passwordHash
	}
	return repos.Admins.Update(ctx, admin)
}
