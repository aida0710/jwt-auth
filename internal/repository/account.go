package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

// accountRepository repository.AccountRepositoryの実装
type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository 新しいアカウントリポジトリを作成
func NewAccountRepository(db *sqlx.DB) domain.AccountRepository {
	return &accountRepository{
		db: db,
	}
}

// Create 新しいアカウントを作成
func (r *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, email, name, created_at, updated_at)
		VALUES (:id, :email, :name, :created_at, :updated_at)
	`

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	_, err := exec.NamedExecContext(ctx, query, account)
	if err != nil {
		return err
	}

	return nil
}

// GetByID IDでアカウントを取得
func (r *accountRepository) GetByID(ctx context.Context, id string) (*domain.Account, error) {
	var account domain.Account
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM accounts
		WHERE id = ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &account, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

// GetByEmail メールアドレスでアカウントを取得
func (r *accountRepository) GetByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var account domain.Account
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM accounts
		WHERE email = ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &account, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}

// List ページネーション付きでアカウント一覧を取得
func (r *accountRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	accounts := make([]*domain.Account, 0)
	query := `
		SELECT id, email, name, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.SelectContext(ctx, &accounts, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

// Update アカウントを更新
func (r *accountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET email = :email, name = :name, updated_at = :updated_at
		WHERE id = :id
	`

	account.UpdatedAt = time.Now()

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.NamedExecContext(ctx, query, account)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

// Delete アカウントを削除
func (r *accountRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM accounts WHERE id = ?`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrAccountNotFound
	}

	return nil
}

// Count アカウントの総数をカウント
func (r *accountRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM accounts`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
