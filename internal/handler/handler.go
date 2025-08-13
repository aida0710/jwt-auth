package handler

import (
	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/usecase"
)

// Server APIサーバーのハンドラー実装
// OpenAPIで生成されたServerInterfaceを実装
type Server struct {
	accountUsecase usecase.AccountUsecase
	projectUsecase usecase.ProjectUsecase
	logger         logger.Logger
}

// NewServer 新しいサーバーインスタンスを作成
func NewServer(
	accountUsecase usecase.AccountUsecase,
	projectUsecase usecase.ProjectUsecase,
	logger logger.Logger,
) api.ServerInterface {
	return &Server{
		accountUsecase: accountUsecase,
		projectUsecase: projectUsecase,
		logger:         logger,
	}
}
