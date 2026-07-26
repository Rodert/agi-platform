package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/javapub/agi-platform-backend/pkg/config"
	"github.com/javapub/agi-platform-backend/pkg/database"
	"github.com/javapub/agi-platform-backend/pkg/logger"
	"gorm.io/gorm"
)

const (
	migrationsDir = "/app/migrations"
	legacyBaseline = "022_expand_gpt_image_2_sizes.sql"
)

func main() {
	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("加载迁移配置失败: %v", err))
	}
	logger.Init(cfg.Server.Debug)
	defer logger.Sync()
	db, err := database.InitMySQL(&cfg.Database)
	if err != nil {
		panic(fmt.Sprintf("连接数据库失败: %v", err))
	}
	defer database.Close()
	if err := applyMigrations(db, migrationsDir); err != nil {
		panic(fmt.Sprintf("执行数据库迁移失败: %v", err))
	}
}

func applyMigrations(db *gorm.DB, dir string) error {
	var locked int
	if err := db.Raw("SELECT GET_LOCK(?, 60)", "agi_platform_migrations").Scan(&locked).Error; err != nil {
		return fmt.Errorf("获取迁移锁失败: %w", err)
	}
	if locked != 1 {
		return fmt.Errorf("等待数据库迁移锁超时")
	}
	defer db.Exec("SELECT RELEASE_LOCK(?)", "agi_platform_migrations")

	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		filename VARCHAR(255) NOT NULL PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='应用数据库迁移记录'`).Error; err != nil {
		return fmt.Errorf("创建迁移记录表失败: %w", err)
	}
	files, err := migrationFiles(dir)
	if err != nil {
		return err
	}
	if err := establishLegacyBaseline(db, files); err != nil {
		return err
	}
	for _, filename := range files {
		var count int64
		if err := db.Table("schema_migrations").Where("filename = ?", filename).Count(&count).Error; err != nil {
			return fmt.Errorf("查询迁移记录 %s 失败: %w", filename, err)
		}
		if count > 0 {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, filename))
		if err != nil {
			return fmt.Errorf("读取迁移 %s 失败: %w", filename, err)
		}
		for _, statement := range strings.Split(string(content), ";") {
			if statement = strings.TrimSpace(statement); statement == "" {
				continue
			}
			if err := db.Exec(statement).Error; err != nil {
				return fmt.Errorf("执行迁移 %s 失败: %w", filename, err)
			}
		}
		if err := db.Exec("INSERT INTO schema_migrations (filename) VALUES (?)", filename).Error; err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %w", filename, err)
		}
		fmt.Printf("已执行迁移: %s\n", filename)
	}
	return nil
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取迁移目录失败: %w", err)
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func establishLegacyBaseline(db *gorm.DB, files []string) error {
	var tracked, modelsTable int64
	if err := db.Table("schema_migrations").Count(&tracked).Error; err != nil {
		return err
	}
	if tracked != 0 {
		return nil
	}
	if err := db.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'ai_models'").Scan(&modelsTable).Error; err != nil {
		return fmt.Errorf("检测历史数据库失败: %w", err)
	}
	if modelsTable == 0 {
		return nil
	}
	for _, filename := range files {
		if filename > legacyBaseline {
			break
		}
		if err := db.Exec("INSERT IGNORE INTO schema_migrations (filename) VALUES (?)", filename).Error; err != nil {
			return fmt.Errorf("建立历史迁移基线失败: %w", err)
		}
	}
	return nil
}
