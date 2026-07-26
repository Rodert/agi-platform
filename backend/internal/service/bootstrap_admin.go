package service

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/javapub/agi-platform-backend/internal/model"
	"github.com/javapub/agi-platform-backend/internal/repository"
	"github.com/javapub/agi-platform-backend/pkg/utils"
	"gorm.io/gorm"
)

// EnsureSuperAdmin creates the configured administrator only when that username
// does not already exist. Changing environment variables never resets an account.
func EnsureSuperAdmin(repo *repository.AdminRepository) error {
	username := strings.TrimSpace(os.Getenv("SUPER_ADMIN_USERNAME"))
	password := os.Getenv("SUPER_ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SUPER_ADMIN_NAME"))
	if username == "" || password == "" {
		return fmt.Errorf("SUPER_ADMIN_USERNAME and SUPER_ADMIN_PASSWORD are required")
	}
	if len(username) < 3 || len(username) > 50 {
		return fmt.Errorf("SUPER_ADMIN_USERNAME must contain 3-50 characters")
	}
	if len(password) < 8 || len(password) > 128 {
		return fmt.Errorf("SUPER_ADMIN_PASSWORD must contain 8-128 characters")
	}
	if name == "" {
		name = "超级管理员"
	}
	if len(name) > 50 {
		return fmt.Errorf("SUPER_ADMIN_NAME must contain at most 50 characters")
	}

	_, err := repo.FindAdminByUsername(username)
	if err == nil {
		return nil
	}
	if err != gorm.ErrRecordNotFound {
		return err
	}

	passwordHash, err := utils.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash super administrator password: %w", err)
	}
	now := time.Now()
	return repo.CreateAdmin(&model.AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		Name:         name,
		Role:         "super_admin",
		Permissions:  "[]",
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
}
