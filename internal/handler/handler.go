package handler

import (
	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/labstack/echo/v4"
)

// Server APIサーバーのハンドラー実装
// OpenAPIで生成されたServerInterfaceを実装
type Server struct {
	accountUsecase usecase.AccountUsecase
	projectUsecase usecase.ProjectUsecase
	authHandler    *AuthHandler
	logger         logger.Logger
}

// NewServer 新しいサーバーインスタンスを作成
func NewServer(
	accountUsecase usecase.AccountUsecase,
	projectUsecase usecase.ProjectUsecase,
	authHandler *AuthHandler,
	logger logger.Logger,
) api.ServerInterface {
	return &Server{
		accountUsecase: accountUsecase,
		projectUsecase: projectUsecase,
		authHandler:    authHandler,
		logger:         logger,
	}
}

// SignUp サインアップエンドポイント
func (s *Server) SignUp(ctx echo.Context) error {
	return s.authHandler.SignUp(ctx)
}

// Login ログインエンドポイント
func (s *Server) Login(ctx echo.Context) error {
	return s.authHandler.Login(ctx)
}

// RefreshToken トークンリフレッシュエンドポイント
func (s *Server) RefreshToken(ctx echo.Context) error {
	return s.authHandler.RefreshToken(ctx)
}

// Logout ログアウトエンドポイント
func (s *Server) Logout(ctx echo.Context) error {
	return s.authHandler.Logout(ctx)
}
