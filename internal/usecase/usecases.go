package usecase

import (
	"context"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/google/uuid"
)

// Usecases すべてのユースケースインターフェースをまとめた構造体
type Usecases struct {
	AccountUsecase AccountUsecase
	ProjectUsecase ProjectUsecase
}

// AccountUsecase アカウントユースケースのインターフェースを定義
type AccountUsecase interface {
	Create(ctx context.Context, input CreateInput) (*domain.Account, error) // SignUpから内部的に使用
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error)
	GetByEmail(ctx context.Context, email string) (*domain.Account, error)
	List(ctx context.Context, limit, offset int) ([]*domain.Account, int, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ProjectUsecase プロジェクトユースケースのインターフェースを定義
type ProjectUsecase interface {
	Create(ctx context.Context, accountID uuid.UUID, input CreateProjectInput) (*domain.Project, error)
	GetByID(ctx context.Context, accountID, projectID uuid.UUID) (*domain.Project, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.Project, int, error)
	Update(ctx context.Context, accountID, projectID uuid.UUID, input UpdateProjectInput) (*domain.Project, error)
	Delete(ctx context.Context, accountID, projectID uuid.UUID) error
}
