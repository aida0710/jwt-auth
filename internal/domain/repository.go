package domain

import (
	"context"
)

// AccountRepository アカウントリポジトリのインターフェースを定義
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	List(ctx context.Context, limit, offset int) ([]*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int, error)
}

// ProjectRepository プロジェクトリポジトリのインターフェースを定義
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id string) (*Project, error)
	GetByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*Project, error)
	List(ctx context.Context, limit, offset int) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id string) error
	DeleteByAccountID(ctx context.Context, accountID string) error
	CountByAccountID(ctx context.Context, accountID string) (int, error)
	Count(ctx context.Context) (int, error)
}
