package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig   `mapstructure:"redis"`
	JWT      JWTConfig     `mapstructure:"jwt"`
	Storage  StorageConfig `mapstructure:"storage"`
	System   SystemConfig  `mapstructure:"system"`
	Worker   WorkerConfig  `mapstructure:"worker"`

	// 注意：Email、Payment、AIModels 配置已移至数据库
	// 对应表: email_config, payment_channels, ai_models
}

type ServerConfig struct {
	Name         string `mapstructure:"name"`
	Env          string `mapstructure:"env"`
	Port         int    `mapstructure:"port"`
	PublicBaseURL string `mapstructure:"public_base_url"`
	Debug        bool   `mapstructure:"debug"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Database     string `mapstructure:"database"`
	Charset      string `mapstructure:"charset"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	LogLevel     string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type StorageConfig struct {
	Type      string `mapstructure:"type"`
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	CDNUrl    string `mapstructure:"cdn_url"`
}

// 注意：EmailConfig、PaymentConfig、AIModelsConfig 已移至数据库
// 对应的数据模型在 internal/model/config.go 和 internal/model/payment.go

type SystemConfig struct {
	MaxUploadSize       int64    `mapstructure:"max_upload_size"`
	AllowedOrigins      []string `mapstructure:"allowed_origins"`
	RateLimitPerMinute  int      `mapstructure:"rate_limit_per_minute"`
}

type WorkerConfig struct {
	Concurrency int    `mapstructure:"concurrency"`
	QueueName   string `mapstructure:"queue_name"`
	RedisStream string `mapstructure:"redis_stream"`
}

var AppConfig *Config

// Load 加载配置
func Load(configPath string) (*Config, error) {
	v := viper.New()

	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		// 默认配置文件路径
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./configs")
		v.AddConfigPath("../configs")
		v.AddConfigPath("../../configs")
	}

	// 环境变量覆盖
	v.AutomaticEnv()

	// 读取配置
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}

	// 环境变量覆盖（支持 .env 文件）
	overrideFromEnv(&config)

	AppConfig = &config
	return &config, nil
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(cfg *Config) {
	if v := os.Getenv("APP_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}
	if v := os.Getenv("PUBLIC_BASE_URL"); v != "" {
		cfg.Server.PublicBaseURL = v
	}
	if v := os.Getenv("DB_HOST"); v != "" {
		cfg.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Database.Port)
	}
	if v := os.Getenv("DB_USER"); v != "" {
		cfg.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		cfg.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		cfg.Database.Database = v
	}
	if v := os.Getenv("REDIS_HOST"); v != "" {
		cfg.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Redis.Port)
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		cfg.JWT.Secret = v
	}
}

// GetDSN 获取数据库连接字符串
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.Charset,
	)
}

// GetRedisAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetJWTExpiration 获取 JWT 过期时间
func (c *JWTConfig) GetExpiration() time.Duration {
	return time.Duration(c.ExpireHours) * time.Hour
}
