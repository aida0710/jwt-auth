package middleware

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// Setup すべてのミドルウェアを設定
func Setup(e *echo.Echo) {
	// エラーハンドラーの初期化
	errorHandler := NewErrorHandler()

	// ロガーの設定
	e.Logger.SetLevel(log.DEBUG)
	e.Logger.SetHeader("Logger: ${time_rfc3339} ${level} [${short_file}:${line}] ${message}")
	e.Logger.SetOutput(os.Stdout)

	// 基本ミドルウェア
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "HttpAccess: time=${time_rfc3339}, method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
		Output: os.Stdout,
	}))
	e.Use(middleware.RecoverWithConfig(errorHandler.RecoverConfig()))
	e.Use(middleware.RequestID())

	// エラーログ出力ミドルウェア
	e.Use(errorHandler.LoggingMiddleware)

	// CORS設定
	e.Use(middleware.CORSWithConfig(getCORSConfig()))

	// タイムアウト設定
	e.Use(middleware.TimeoutWithConfig(getTimeoutConfig()))

	// GZIP圧縮
	e.Use(middleware.Gzip())

	// セキュリティヘッダー
	e.Use(middleware.SecureWithConfig(getSecureConfig()))

	// カスタムエラーハンドラー
	e.HTTPErrorHandler = errorHandler.HTTPErrorHandler
}

// getCORSConfig CORS設定を返す
func getCORSConfig() middleware.CORSConfig {
	return middleware.DefaultCORSConfig
}

// getTimeoutConfig タイムアウト設定を返す
func getTimeoutConfig() middleware.TimeoutConfig {
	return middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}
}

// getSecureConfig セキュリティヘッダー設定を返す
func getSecureConfig() middleware.SecureConfig {
	return middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'",
	}
}
