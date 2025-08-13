package repository

import (
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/jmoiron/sqlx"
)

// Repositories すべてのリポジトリを集約するインターフェース
type Repositories interface {
	Account() domain.AccountRepository
	Project() domain.ProjectRepository
}

// repositories 実装構造体
type repositories struct {
	account domain.AccountRepository
	project domain.ProjectRepository
}

// NewRepositories リポジトリ集約を生成
func NewRepositories(db *sqlx.DB) Repositories {
	return &repositories{
		account: NewAccountRepository(db),
		project: NewProjectRepository(db),
	}
}

// Account アカウントリポジトリを返す
func (r *repositories) Account() domain.AccountRepository {
	return r.account
}

// Project プロジェクトリポジトリを返す
func (r *repositories) Project() domain.ProjectRepository {
	return r.project
}
