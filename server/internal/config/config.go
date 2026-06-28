package config

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Auth     AuthConfig
	Admin    AdminConfig
	Storage  StorageConfig
}

type AppConfig struct {
	Env  string
	Name string
}

type HTTPConfig struct {
	Host string
	Port int
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	Charset      string
	ParseTime    bool
	Loc          string
	MaxIdleConns int
	MaxOpenConns int
	MaxLifetime  time.Duration
}

type AuthConfig struct {
	JWTSecret           string
	TokenLifetime       time.Duration
	RegisterGiftCredits int64
}

type AdminConfig struct {
	Enabled                bool
	Username               string
	Password               string
	ResetPasswordOnStartup bool
	Nickname               string
	Role                   string
	Status                 string
}

type StorageConfig struct {
	Provider      string
	LocalRoot     string
	COS           COSConfig
	SecretCSVPath string
}

type COSConfig struct {
	SecretID      string
	SecretKey     string
	Bucket        string
	Region        string
	PublicBaseURL string
	UploadPrefix  string
}

type rawConfig struct {
	App struct {
		Env  string `yaml:"env"`
		Name string `yaml:"name"`
	} `yaml:"app"`
	HTTP struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"http"`
	Database struct {
		Host                   string `yaml:"host"`
		Port                   int    `yaml:"port"`
		User                   string `yaml:"user"`
		Password               string `yaml:"password"`
		Name                   string `yaml:"name"`
		Charset                string `yaml:"charset"`
		ParseTime              *bool  `yaml:"parse_time"`
		Loc                    string `yaml:"loc"`
		MaxIdleConns           int    `yaml:"max_idle_conns"`
		MaxOpenConns           int    `yaml:"max_open_conns"`
		ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"`
	} `yaml:"database"`
	Auth struct {
		JWTSecret           string `yaml:"jwt_secret"`
		JWTExpireSeconds    int    `yaml:"jwt_expire_seconds"`
		RegisterGiftCredits int64  `yaml:"register_gift_credits"`
	} `yaml:"auth"`
	Admin struct {
		Enabled                *bool  `yaml:"enabled"`
		Username               string `yaml:"username"`
		Password               string `yaml:"password"`
		ResetPasswordOnStartup *bool  `yaml:"reset_password_on_startup"`
		Nickname               string `yaml:"nickname"`
		Role                   string `yaml:"role"`
		Status                 string `yaml:"status"`
	} `yaml:"admin"`
	Storage struct {
		Provider      string `yaml:"provider"`
		LocalRoot     string `yaml:"local_root"`
		SecretCSVPath string `yaml:"secret_csv_path"`
		COS           struct {
			SecretID      string `yaml:"secret_id"`
			SecretKey     string `yaml:"secret_key"`
			Bucket        string `yaml:"bucket"`
			Region        string `yaml:"region"`
			PublicBaseURL string `yaml:"public_base_url"`
			UploadPrefix  string `yaml:"upload_prefix"`
		} `yaml:"cos"`
	} `yaml:"storage"`
}

func (c DatabaseConfig) DSN() string {
	parseTime := "False"
	if c.ParseTime {
		parseTime = "True"
	}
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%s&loc=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.Charset,
		parseTime,
		c.Loc,
	)
}

func Load() Config {
	raw := readRawConfig(configPath())
	cfg := Config{
		App: AppConfig{
			Env:  firstNonEmpty(raw.App.Env, "local"),
			Name: firstNonEmpty(raw.App.Name, "agi-platform"),
		},
		HTTP: HTTPConfig{
			Host: firstNonEmpty(raw.HTTP.Host, "0.0.0.0"),
			Port: firstNonZeroInt(raw.HTTP.Port, 8080),
		},
		Database: DatabaseConfig{
			Host:         firstNonEmpty(raw.Database.Host, "127.0.0.1"),
			Port:         firstNonZeroInt(raw.Database.Port, 3306),
			User:         firstNonEmpty(raw.Database.User, "root"),
			Password:     raw.Database.Password,
			Name:         firstNonEmpty(raw.Database.Name, "agi_platform"),
			Charset:      firstNonEmpty(raw.Database.Charset, "utf8mb4"),
			ParseTime:    firstBool(raw.Database.ParseTime, true),
			Loc:          firstNonEmpty(raw.Database.Loc, "Local"),
			MaxIdleConns: firstNonZeroInt(raw.Database.MaxIdleConns, 10),
			MaxOpenConns: firstNonZeroInt(raw.Database.MaxOpenConns, 50),
			MaxLifetime:  time.Duration(firstNonZeroInt(raw.Database.ConnMaxLifetimeSeconds, 3600)) * time.Second,
		},
		Auth: AuthConfig{
			JWTSecret:           firstNonEmpty(raw.Auth.JWTSecret, "local-dev-secret-change-me"),
			TokenLifetime:       time.Duration(firstNonZeroInt(raw.Auth.JWTExpireSeconds, 604800)) * time.Second,
			RegisterGiftCredits: firstNonZeroInt64(raw.Auth.RegisterGiftCredits, 100),
		},
		Admin: AdminConfig{
			Enabled:                firstBool(raw.Admin.Enabled, true),
			Username:               firstNonEmpty(raw.Admin.Username, "admin"),
			Password:               raw.Admin.Password,
			ResetPasswordOnStartup: firstBool(raw.Admin.ResetPasswordOnStartup, false),
			Nickname:               firstNonEmpty(raw.Admin.Nickname, "Administrator"),
			Role:                   firstNonEmpty(raw.Admin.Role, "super_admin"),
			Status:                 firstNonEmpty(raw.Admin.Status, "active"),
		},
		Storage: StorageConfig{
			Provider:      firstNonEmpty(raw.Storage.Provider, "cos"),
			LocalRoot:     firstNonEmpty(raw.Storage.LocalRoot, "uploads"),
			SecretCSVPath: firstNonEmpty(raw.Storage.SecretCSVPath, "../SecretKey.csv"),
			COS: COSConfig{
				SecretID:      raw.Storage.COS.SecretID,
				SecretKey:     raw.Storage.COS.SecretKey,
				Bucket:        firstNonEmpty(raw.Storage.COS.Bucket, "agi-platform-dev-1257142189"),
				Region:        firstNonEmpty(raw.Storage.COS.Region, "ap-guangzhou"),
				PublicBaseURL: firstNonEmpty(raw.Storage.COS.PublicBaseURL, "https://agi-platform-dev-1257142189.cos.ap-guangzhou.myqcloud.com"),
				UploadPrefix:  raw.Storage.COS.UploadPrefix,
			},
		},
	}
	applyEnvOverrides(&cfg)
	applySecretCSV(&cfg)
	return cfg
}

func readRawConfig(path string) rawConfig {
	var raw rawConfig
	if path == "" {
		return raw
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return raw
	}
	_ = yaml.Unmarshal(data, &raw)
	return raw
}

func configPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	for _, path := range []string{"config.yaml", "../config.yaml", "/app/config.yaml"} {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func applyEnvOverrides(cfg *Config) {
	cfg.App.Env = getEnv("APP_ENV", cfg.App.Env)
	cfg.App.Name = getEnv("APP_NAME", cfg.App.Name)
	cfg.HTTP.Host = getEnv("HTTP_HOST", cfg.HTTP.Host)
	cfg.HTTP.Port = getEnvInt("HTTP_PORT", cfg.HTTP.Port)
	cfg.Database.Host = getEnv("MYSQL_HOST", cfg.Database.Host)
	cfg.Database.Port = getEnvInt("MYSQL_PORT", cfg.Database.Port)
	cfg.Database.User = getEnv("MYSQL_USER", cfg.Database.User)
	cfg.Database.Password = getEnv("MYSQL_PASSWORD", cfg.Database.Password)
	cfg.Database.Name = getEnv("MYSQL_DATABASE", cfg.Database.Name)
	cfg.Database.Charset = getEnv("MYSQL_CHARSET", cfg.Database.Charset)
	cfg.Database.ParseTime = getEnvBool("MYSQL_PARSE_TIME", cfg.Database.ParseTime)
	cfg.Database.Loc = getEnv("MYSQL_LOC", cfg.Database.Loc)
	cfg.Database.MaxIdleConns = getEnvInt("MYSQL_MAX_IDLE_CONNS", cfg.Database.MaxIdleConns)
	cfg.Database.MaxOpenConns = getEnvInt("MYSQL_MAX_OPEN_CONNS", cfg.Database.MaxOpenConns)
	cfg.Database.MaxLifetime = time.Duration(getEnvInt("MYSQL_CONN_MAX_LIFETIME_SECONDS", int(cfg.Database.MaxLifetime.Seconds()))) * time.Second
	cfg.Auth.JWTSecret = getEnv("JWT_SECRET", cfg.Auth.JWTSecret)
	cfg.Auth.TokenLifetime = time.Duration(getEnvInt("JWT_EXPIRE_SECONDS", int(cfg.Auth.TokenLifetime.Seconds()))) * time.Second
	cfg.Auth.RegisterGiftCredits = int64(getEnvInt("REGISTER_GIFT_CREDITS", int(cfg.Auth.RegisterGiftCredits)))
	cfg.Admin.Enabled = getEnvBool("ADMIN_BOOTSTRAP_ENABLED", cfg.Admin.Enabled)
	cfg.Admin.Username = getEnv("ADMIN_USERNAME", cfg.Admin.Username)
	cfg.Admin.Password = getEnv("ADMIN_PASSWORD", cfg.Admin.Password)
	cfg.Admin.ResetPasswordOnStartup = getEnvBool("ADMIN_RESET_PASSWORD_ON_STARTUP", cfg.Admin.ResetPasswordOnStartup)
	cfg.Admin.Nickname = getEnv("ADMIN_NICKNAME", cfg.Admin.Nickname)
	cfg.Admin.Role = getEnv("ADMIN_ROLE", cfg.Admin.Role)
	cfg.Admin.Status = getEnv("ADMIN_STATUS", cfg.Admin.Status)
	cfg.Storage.Provider = getEnv("STORAGE_PROVIDER", cfg.Storage.Provider)
	cfg.Storage.LocalRoot = getEnv("STORAGE_LOCAL_ROOT", cfg.Storage.LocalRoot)
	cfg.Storage.SecretCSVPath = getEnv("COS_SECRET_CSV_PATH", cfg.Storage.SecretCSVPath)
	cfg.Storage.COS.SecretID = getEnv("COS_SECRET_ID", cfg.Storage.COS.SecretID)
	cfg.Storage.COS.SecretKey = getEnv("COS_SECRET_KEY", cfg.Storage.COS.SecretKey)
	cfg.Storage.COS.Bucket = getEnv("COS_BUCKET", cfg.Storage.COS.Bucket)
	cfg.Storage.COS.Region = getEnv("COS_REGION", cfg.Storage.COS.Region)
	cfg.Storage.COS.PublicBaseURL = getEnv("COS_PUBLIC_BASE_URL", cfg.Storage.COS.PublicBaseURL)
	cfg.Storage.COS.UploadPrefix = getEnv("COS_UPLOAD_PREFIX", cfg.Storage.COS.UploadPrefix)
}

func applySecretCSV(cfg *Config) {
	if cfg.Storage.COS.SecretID != "" && cfg.Storage.COS.SecretKey != "" {
		return
	}
	secretID, secretKey := readSecretCSV(cfg.Storage.SecretCSVPath)
	if cfg.Storage.COS.SecretID == "" {
		cfg.Storage.COS.SecretID = secretID
	}
	if cfg.Storage.COS.SecretKey == "" {
		cfg.Storage.COS.SecretKey = secretKey
	}
}

func readSecretCSV(primaryPath string) (string, string) {
	paths := []string{primaryPath, "../SecretKey.csv", "SecretKey.csv", "/app/SecretKey.csv"}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		rows, err := csv.NewReader(file).ReadAll()
		_ = file.Close()
		if err != nil || len(rows) < 2 || len(rows[1]) < 2 {
			continue
		}
		return strings.TrimSpace(rows[1][0]), strings.TrimSpace(rows[1][1])
	}
	return "", ""
}

func firstNonEmpty(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func firstNonZeroInt(value int, fallback int) int {
	if value != 0 {
		return value
	}
	return fallback
}

func firstNonZeroInt64(value int64, fallback int64) int64 {
	if value != 0 {
		return value
	}
	return fallback
}

func firstBool(value *bool, fallback bool) bool {
	if value == nil {
		return fallback
	}
	return *value
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}
