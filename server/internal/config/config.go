package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App      AppConfig
	HTTP     HTTPConfig
	Database DatabaseConfig
	Auth     AuthConfig
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
	return Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "local"),
			Name: getEnv("APP_NAME", "agi-platform"),
		},
		HTTP: HTTPConfig{
			Host: getEnv("HTTP_HOST", "0.0.0.0"),
			Port: getEnvInt("HTTP_PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:         getEnv("MYSQL_HOST", "127.0.0.1"),
			Port:         getEnvInt("MYSQL_PORT", 3306),
			User:         getEnv("MYSQL_USER", "root"),
			Password:     getEnv("MYSQL_PASSWORD", ""),
			Name:         getEnv("MYSQL_DATABASE", "agi_platform"),
			Charset:      getEnv("MYSQL_CHARSET", "utf8mb4"),
			ParseTime:    getEnvBool("MYSQL_PARSE_TIME", true),
			Loc:          getEnv("MYSQL_LOC", "Local"),
			MaxIdleConns: getEnvInt("MYSQL_MAX_IDLE_CONNS", 10),
			MaxOpenConns: getEnvInt("MYSQL_MAX_OPEN_CONNS", 50),
			MaxLifetime:  time.Duration(getEnvInt("MYSQL_CONN_MAX_LIFETIME_SECONDS", 3600)) * time.Second,
		},
		Auth: AuthConfig{
			JWTSecret:           getEnv("JWT_SECRET", "local-dev-secret-change-me"),
			TokenLifetime:       time.Duration(getEnvInt("JWT_EXPIRE_SECONDS", 604800)) * time.Second,
			RegisterGiftCredits: int64(getEnvInt("REGISTER_GIFT_CREDITS", 100)),
		},
	}
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
