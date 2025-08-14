package usecase

import (
	"context"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/google/uuid"
)

// CreateProjectInput プロジェクト作成用の入力
type CreateProjectInput struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	Status      *string `json:"status,omitempty"`
}

// UpdateProjectInput プロジェクト更新用の入力
type UpdateProjectInput struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Status      *string `json:"status,omitempty"`
}

// projectUsecase ProjectUsecaseインターフェースの実装
type projectUsecase struct {
	projectRepo domain.ProjectRepository
	accountRepo domain.AccountRepository
	txManager   database.TransactionManager
}

// NewProjectUsecase 新しいプロジェクトユースケースを作成
func NewProjectUsecase(
	projectRepo domain.ProjectRepository,
	accountRepo domain.AccountRepository,
	txManager database.TransactionManager,
) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
		accountRepo: accountRepo,
		txManager:   txManager,
	}
}

// Create 新しいプロジェクトを作成
func (u *projectUsecase) Create(ctx context.Context, accountID uuid.UUID, input CreateProjectInput) (*domain.Project, error) {
	// アカウントが存在するか確認
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	// プロジェクト数の制限をチェック
	projects, err := u.projectRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if len(projects) >= domain.MaxProjectsPerAccount {
		return nil, domain.ErrProjectLimitExceeded
	}

	// Domain層のファクトリメソッドを使用
	project := domain.NewProject(accountID, input.Name, input.Description)

	// ステータスの処理を文字列として統一
	if input.Status != nil {
		project.Status = domain.ProjectStatus(*input.Status)
	} else {
		project.Status = domain.ProjectStatusActive
	}

	if err := project.Validate(); err != nil {
		return nil, err
	}

	if !project.IsValidStatus() {
		return nil, domain.ErrInvalidStatus
	}

	if err := u.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

// GetByID IDでプロジェクトを取得
func (u *projectUsecase) GetByID(ctx context.Context, accountID, projectID uuid.UUID) (*domain.Project, error) {
	// Verify account exists
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, domain.ErrProjectNotFound
	}

	// プロジェクトがアカウントに属しているか確認
	if project.AccountID != accountID {
		return nil, domain.ErrProjectNotFound
	}

	return project, nil
}

// ListByAccountID アカウントIDでプロジェクト一覧を取得
func (u *projectUsecase) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*domain.Project, error) {
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	projects, err := u.projectRepo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Update プロジェクトを更新
func (u *projectUsecase) Update(ctx context.Context, accountID, projectID uuid.UUID, input UpdateProjectInput) (*domain.Project, error) {
	var updatedProject *domain.Project

	// トランザクション内で実行
	err := u.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		// Verify account exists
		account, err := u.accountRepo.GetByID(ctx, accountID)
		if err != nil {
			return err
		}
		if account == nil {
			return domain.ErrAccountNotFound
		}

		project, err := u.projectRepo.GetByID(ctx, projectID)
		if err != nil {
			return err
		}
		if project == nil {
			return domain.ErrProjectNotFound
		}

		// Verify the project belongs to the account
		if project.AccountID != accountID {
			return domain.ErrProjectNotFound
		}

		if input.Name != nil {
			project.Name = *input.Name
		}

		if input.Description != nil {
			project.Description = *input.Description
		}

		if input.Status != nil {
			project.Status = domain.ProjectStatus(*input.Status)
			if !project.IsValidStatus() {
				return domain.ErrInvalidStatus
			}
		}

		if err := project.Validate(); err != nil {
			return err
		}

		if err := u.projectRepo.Update(ctx, project); err != nil {
			return err
		}

		updatedProject = project
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedProject, nil
}

// Delete プロジェクトを削除
func (u *projectUsecase) Delete(ctx context.Context, accountID, projectID uuid.UUID) error {
	// Verify account exists
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return err
	}
	if account == nil {
		return domain.ErrAccountNotFound
	}

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}
	if project == nil {
		return domain.ErrProjectNotFound
	}

	// Verify the project belongs to the account
	if project.AccountID != accountID {
		return domain.ErrProjectNotFound
	}

	if err := u.projectRepo.Delete(ctx, projectID); err != nil {
		return err
	}

	return nil
}
