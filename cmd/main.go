package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/config"
	"github.com/aida0710/jwt-auth/internal/di"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/middleware"
	"github.com/labstack/echo/v4"
)

func main() {
	// 設定の読み込み
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// DIコンテナの初期化
	container, err := di.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize container: %v", err)
	}
	defer func(container *di.Container) {
		err := container.Close()
		if err != nil {
			log.Fatalf("Failed to close container: %v", err)
			return
		}
	}(container)

	// Echoインスタンスの作成
	e := echo.New()

	// すべてのミドルウェアを設定
	middleware.Setup(e)

	// OpenAPIハンドラーの登録
	// baseURLに/api/v1を指定
	api.RegisterHandlersWithBaseURL(e, container.GetHandler(), "/api/v1")

	// ヘルスチェックエンドポイント
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"service": "JWT Auth API",
			"status":  "running",
			"version": "1.0.0",
		})
	})

	// サーバーの起動
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// グレースフルシャットダウン
	go func() {
		container.GetLogger().Info(context.Background(), "Starting server",
			logger.F("port", cfg.Server.Port),
			logger.F("env", cfg.Env),
		)

		if err := e.StartServer(srv); err != nil && !errors.Is(err, http.ErrServerClosed) {
			container.GetLogger().Fatal(context.Background(), "Failed to start server", err)
		}
	}()

	// シグナル待機
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// グレースフルシャットダウンの実行
	container.GetLogger().Info(context.Background(), "Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		container.GetLogger().Error(context.Background(), "Failed to shutdown server", err)
	}

	container.GetLogger().Info(context.Background(), "Server exited")
}
