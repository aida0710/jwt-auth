package handler

import (
	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/labstack/echo/v4"
)

// Handler すべてのハンドラーを集約するインターフェース
type Handler interface {
	AccountHandler
	ProjectHandler
	HealthHandler
}

// AccountHandler アカウント関連のハンドラーインターフェース
type AccountHandler interface {
	// ListAccounts アカウント一覧取得
	ListAccounts(ctx echo.Context, params []api.Account) error
	// CreateAccount アカウント作成
	CreateAccount(ctx echo.Context) error
	// GetAccount アカウント取得
	GetAccount(ctx echo.Context, accountId api.AccountID) error
	// UpdateAccount アカウント更新
	UpdateAccount(ctx echo.Context, accountId api.AccountID) error
	// DeleteAccount アカウント削除
	DeleteAccount(ctx echo.Context, accountId api.AccountID) error
}

// ProjectHandler プロジェクト関連のハンドラーインターフェース
type ProjectHandler interface {
	// ListProjects プロジェクト一覧取得
	ListProjects(ctx echo.Context, accountId api.AccountID, params []api.Project) error
	// CreateProject プロジェクト作成
	CreateProject(ctx echo.Context, accountId api.AccountID) error
	// GetProject プロジェクト取得
	GetProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error
	// UpdateProject プロジェクト更新
	UpdateProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error
	// DeleteProject プロジェクト削除
	DeleteProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error
}

// HealthHandler ヘルスチェック関連のハンドラーインターフェース
type HealthHandler interface {
	// GetHealth ヘルスチェック
	GetHealth(ctx echo.Context) error
}
