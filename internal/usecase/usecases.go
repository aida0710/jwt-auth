package usecase

import (
	"context"

	"github.com/aida0710/jwt-auth/internal/domain"
)

// Usecases すべてのユースケースインターフェースをまとめた構造体
type Usecases struct {
	AccountUsecase AccountUsecase
	ProjectUsecase ProjectUsecase
}

// AccountUsecase アカウントユースケースのインターフェースを定義
type AccountUsecase interface {
	Create(ctx context.Context, input CreateInput) (*domain.Account, error)
	GetByID(ctx context.Context, id string) (*domain.Account, error)
	GetByEmail(ctx context.Context, email string) (*domain.Account, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Account, int, error)
	Update(ctx context.Context, id string, input UpdateInput) (*domain.Account, error)
	Delete(ctx context.Context, id string) error
}

// ProjectUsecase プロジェクトユースケースのインターフェースを定義
type ProjectUsecase interface {
	Create(ctx context.Context, accountID string, input CreateProjectInput) (*domain.Project, error)
	GetByID(ctx context.Context, accountID, projectID string) (*domain.Project, error)
	ListByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*domain.Project, int, error)
	Update(ctx context.Context, accountID, projectID string, input UpdateProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, accountID, projectID string) error
}
