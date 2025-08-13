package usecase

import (
	"context"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/google/uuid"
)

// CreateInput アカウント作成用の入力
type CreateInput struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required"`
}

// UpdateInput アカウント更新用の入力
type UpdateInput struct {
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
	Name  *string `json:"name,omitempty"`
}

// accountUsecase AccountUsecaseインターフェースの実装
type accountUsecase struct {
	accountRepo domain.AccountRepository
	projectRepo domain.ProjectRepository
	txManager   database.TransactionManager
}

// NewAccountUsecase 新しいアカウントユースケースを作成
func NewAccountUsecase(
	accountRepo domain.AccountRepository,
	projectRepo domain.ProjectRepository,
	txManager database.TransactionManager,
) AccountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
		projectRepo: projectRepo,
		txManager:   txManager,
	}
}

// Create 新しいアカウントを作成
func (u *accountUsecase) Create(ctx context.Context, input CreateInput) (*domain.Account, error) {
	existing, _ := u.accountRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, domain.ErrDuplicateEmail
	}

	account := &domain.Account{
		ID:    uuid.New().String(),
		Email: input.Email,
		Name:  input.Name,
	}

	if err := account.Validate(); err != nil {
		return nil, err
	}

	if err := u.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetByID IDでアカウントを取得
func (u *accountUsecase) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	if err := domain.ValidateAccountID(id); err != nil {
		return nil, err
	}

	account, err := u.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	return account, nil
}

// GetByEmail メールアドレスでアカウントを取得
func (u *accountUsecase) GetByEmail(ctx context.Context, email string) (*domain.Account, error) {
	if email == "" {
		return nil, domain.ErrInvalidEmail
	}

	account, err := u.accountRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	return account, nil
}

// List ページネーション付きでアカウント一覧を取得
func (u *accountUsecase) List(ctx context.Context, limit, offset int) ([]*domain.Account, int, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	accounts, err := u.accountRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.accountRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return accounts, total, nil
}

// Update アカウントを更新
func (u *accountUsecase) Update(ctx context.Context, id string, input UpdateInput) (*domain.Account, error) {
	if err := domain.ValidateAccountID(id); err != nil {
		return nil, err
	}

	account, err := u.accountRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, domain.ErrAccountNotFound
	}

	if input.Email != nil && *input.Email != account.Email {
		existing, _ := u.accountRepo.GetByEmail(ctx, *input.Email)
		if existing != nil {
			return nil, domain.ErrDuplicateEmail
		}
		account.Email = *input.Email
	}

	if input.Name != nil {
		account.Name = *input.Name
	}

	if err := account.Validate(); err != nil {
		return nil, err
	}

	if err := u.accountRepo.Update(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// Delete アカウントとそのプロジェクトを削除
func (u *accountUsecase) Delete(ctx context.Context, id string) error {
	if err := domain.ValidateAccountID(id); err != nil {
		return err
	}

	return u.txManager.RunInTransaction(ctx, func(ctx context.Context) error {
		account, err := u.accountRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if account == nil {
			return domain.ErrAccountNotFound
		}

		// このアカウントに関連するすべてのプロジェクトを削除
		if err := u.projectRepo.DeleteByAccountID(ctx, id); err != nil {
			return err
		}

		// アカウントを削除
		if err := u.accountRepo.Delete(ctx, id); err != nil {
			return err
		}

		return nil
	})
}
