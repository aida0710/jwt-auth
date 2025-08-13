package di

import (
	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/config"
	"github.com/aida0710/jwt-auth/internal/handler"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/repository"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/jmoiron/sqlx"
)

// Container DIコンテナの構造体
type Container struct {
	config    *config.Config
	db        *sqlx.DB
	logger    logger.Logger
	txManager database.TransactionManager
	repos     repository.Repositories
	handler   api.ServerInterface
}

// NewContainer 新しいDIコンテナを作成
func NewContainer(cfg *config.Config) (*Container, error) {
	// データベース接続の初期化
	dbConfig := &database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Database,
	}

	db, err := database.NewMySQLConnection(dbConfig)
	if err != nil {
		return nil, err
	}

	// コネクションプールの設定
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// ロガーの初期化
	log := logger.NewLogger(cfg.Logger.Level, cfg.Logger.Format)

	// トランザクションマネージャーの初期化
	txManager := database.NewTransactionManager(db)

	// リポジトリの初期化
	repos := repository.NewRepositories(db)

	// ユースケースの初期化
	accountUsecase := usecase.NewAccountUsecase(
		repos.Account(),
		repos.Project(),
		txManager,
	)
	projectUsecase := usecase.NewProjectUsecase(
		repos.Project(),
		repos.Account(),
		txManager,
	)

	// ハンドラーの初期化
	h := handler.NewServer(
		accountUsecase,
		projectUsecase,
		log,
	)

	return &Container{
		config:    cfg,
		db:        db,
		logger:    log,
		txManager: txManager,
		repos:     repos,
		handler:   h,
	}, nil
}

// Close コンテナのリソースをクリーンアップ
func (c *Container) Close() error {
	return c.DB().Close()
}

// GetLogger ロガーを返す
func (c *Container) GetLogger() logger.Logger {
	return c.logger
}

// GetTxManager トランザクションマネージャーを返す
func (c *Container) GetTxManager() database.TransactionManager {
	return c.txManager
}

// GetRepositories リポジトリを返す
func (c *Container) GetRepositories() repository.Repositories {
	return c.repos
}

// GetHandler ハンドラーを返す（OpenAPIのServerInterfaceを返す）
func (c *Container) GetHandler() api.ServerInterface {
	return c.handler
}

// DB データベース接続を返す
func (c *Container) DB() *sqlx.DB {
	return c.db
}
