package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Auth     AuthConfig
	Storage  StorageConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Driver      string
	SQLitePath  string
	PostgresDSN string `mapstructure:"postgres_dsn"`
}

type RedisConfig struct {
	Enabled   bool
	Addr      string
	Password  string
	DB        int
	StreamKey string `mapstructure:"stream_key"`
	CacheTTL  int    `mapstructure:"cache_ttl"` // seconds
}

type AuthConfig struct {
	Enabled    bool
	AdminToken string `mapstructure:"admin_token"`
}

type StorageConfig struct {
	Driver        string // local | minio
	LocalPath     string `mapstructure:"local_path"`
	Prefix        string // 资源根前缀，local/minio 目录规则一致，默认 uploads
	PublicBaseURL string `mapstructure:"public_base_url"`
	MinIO         MinIOConfig `mapstructure:"minio"`
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string `mapstructure:"access_key"`
	SecretKey  string `mapstructure:"secret_key"`
	Bucket     string
	UseSSL     bool   `mapstructure:"use_ssl"`
	Prefix     string // bucket 内根目录，如 uploads
	PublicRead bool   `mapstructure:"public_read"` // 允许匿名读取对象（商品图/视频外链）
}

type CORSConfig struct {
	AllowOrigins []string
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8088
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "postgres"
	}
	if cfg.Database.SQLitePath == "" {
		cfg.Database.SQLitePath = "./data/productcore.db"
	}
	if cfg.Database.PostgresDSN == "" {
		cfg.Database.PostgresDSN = "host=127.0.0.1 user=postgres password=postgres dbname=productcore port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	if cfg.Redis.StreamKey == "" {
		cfg.Redis.StreamKey = "productcore:product:events"
	}
	if cfg.Redis.CacheTTL == 0 {
		cfg.Redis.CacheTTL = 300
	}
	if cfg.Storage.Driver == "" {
		cfg.Storage.Driver = "local"
	}
	if cfg.Storage.LocalPath == "" {
		cfg.Storage.LocalPath = "./data/uploads"
	}
	if cfg.Storage.PublicBaseURL == "" {
		cfg.Storage.PublicBaseURL = "http://localhost:8090/uploads"
	}
	if cfg.Storage.MinIO.Prefix == "" {
		cfg.Storage.MinIO.Prefix = "uploads"
	}
	if !v.IsSet("storage.minio.public_read") {
		cfg.Storage.MinIO.PublicRead = true
	}
	if cfg.Auth.AdminToken == "" {
		cfg.Auth.AdminToken = "dev-admin-token"
	}
	return &cfg, nil
}
