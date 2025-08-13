package database

import (
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Config データベース設定
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

// NewMySQLConnection デフォルト設定で新しいMySQL接続を作成
func NewMySQLConnection(cfg *Config) (*sqlx.DB, error) {
	// デフォルト値
	charset := "utf8mb4"
	parseTime := true
	loc := "Local"
	maxOpen := 25
	maxIdle := 25
	lifetime := 5 * time.Minute

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		charset,
		parseTime,
		loc,
	)

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(lifetime)

	// 接続を確認するためにデータベースにPing
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
