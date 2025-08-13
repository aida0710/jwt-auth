package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config アプリケーション全体の設定を保持
type Config struct {
	Env      string
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

// ServerConfig サーバー関連の設定
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DatabaseConfig データベース関連の設定
type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// JWTConfig JWT関連の設定
type JWTConfig struct {
	Secret          string
	ExpireDuration  time.Duration
	RefreshDuration time.Duration
}

// LoggerConfig ロガー関連の設定
type LoggerConfig struct {
	Level  string
	Format string // jsonまたはtext
}

// LoadConfig 環境変数から設定を読み込む
func LoadConfig() (*Config, error) {
	// .envファイルが存在する場合は読み込む
	_ = godotenv.Load()

	config := &Config{
		Env: getEnv("APP_ENV", "development"),
		Server: ServerConfig{
			Port:         getEnv("BACKEND_PORT", "8080"),
			ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:  getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getIntEnv("DB_PORT", 3306),
			User:            getEnv("DB_USER", "root"),
			Password:        getEnv("DB_PASSWORD", "password"),
			Database:        getEnv("DB_NAME", "jwt_auth"),
			MaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
			ExpireDuration:  getDurationEnv("JWT_EXPIRE_DURATION", 1*time.Hour),
			RefreshDuration: getDurationEnv("JWT_REFRESH_DURATION", 24*time.Hour*7),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}

	// 必須項目のバリデーション
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

// Validate 設定の妥当性を検証
func (c *Config) Validate() error {
	if c.Database.Password == "" && c.Env == "production" {
		return fmt.Errorf("DB_PASSWORD is required in production environment")
	}

	if c.JWT.Secret == "your-secret-key-change-this-in-production" && c.Env == "production" {
		return fmt.Errorf("JWT_SECRET must be changed in production environment")
	}

	return nil
}

// IsDevelopment 開発環境かどうかを返す
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsProduction 本番環境かどうかを返す
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// getEnv 環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getIntEnv 環境変数を整数として取得
func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getDurationEnv 環境変数を時間として取得
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
