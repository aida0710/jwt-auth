package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// accountDB データベース用のアカウント構造体（UUIDをstringで保存）
type accountDB struct {
	ID           string    `db:"id"`
	Email        string    `db:"email"`
	Name         string    `db:"name"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// toDomain DB構造体からドメインモデルへ変換
func (a *accountDB) toDomain() (*domain.Account, error) {
	id, err := uuid.Parse(a.ID)
	if err != nil {
		return nil, err
	}

	return &domain.Account{
		ID:           id,
		Email:        a.Email,
		Name:         a.Name,
		PasswordHash: a.PasswordHash,
		CreatedAt:    a.CreatedAt,
		UpdatedAt:    a.UpdatedAt,
	}, nil
}

// fromDomain ドメインモデルからDB構造体へ変換
func fromDomainAccount(account *domain.Account) *accountDB {
	return &accountDB{
		ID:           account.ID.String(),
		Email:        account.Email,
		Name:         account.Name,
		PasswordHash: account.PasswordHash,
		CreatedAt:    account.CreatedAt,
		UpdatedAt:    account.UpdatedAt,
	}
}

// accountRepository repository.AccountRepositoryの実装
type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository アカウントリポジトリを作成
func NewAccountRepository(db *sqlx.DB) domain.AccountRepository {
	return &accountRepository{
		db: db,
	}
}

// Create 新しいアカウントを作成
func (r *accountRepository) Create(ctx context.Context, account *domain.Account) error {
	query := `
		INSERT INTO accounts (id, email, name, password_hash, created_at, updated_at)
		VALUES (:id, :email, :name, :password_hash, :created_at, :updated_at)
	`

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	dbAccount := fromDomainAccount(account)

	exec := database.GetExecutor(ctx, r.db)
	_, err := exec.NamedExecContext(ctx, query, dbAccount)
	if err != nil {
		return err
	}

	return nil
}

// GetByID IDでアカウントを取得
func (r *accountRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Account, error) {
	var dbAccount accountDB
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM accounts
		WHERE id = ?
	`

	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &dbAccount, query, id.String())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return dbAccount.toDomain()
}

// GetByEmail メールアドレスでアカウントを取得
func (r *accountRepository) GetByEmail(ctx context.Context, email string) (*domain.Account, error) {
	var dbAccount accountDB
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM accounts
		WHERE email = ?
	`

	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &dbAccount, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return dbAccount.toDomain()
}

// List ページネーション付きでアカウント一覧を取得
func (r *accountRepository) List(ctx context.Context, limit, offset int) ([]*domain.Account, error) {
	dbAccounts := make([]accountDB, 0)
	query := `
		SELECT id, email, name, password_hash, created_at, updated_at
		FROM accounts
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	exec := database.GetExecutor(ctx, r.db)
	err := exec.SelectContext(ctx, &dbAccounts, query, limit, offset)
	if err != nil {
		return nil, err
	}

	accounts := make([]*domain.Account, 0, len(dbAccounts))
	for _, dbAcc := range dbAccounts {
		acc, err := dbAcc.toDomain()
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}

	return accounts, nil
}

// Update アカウントを更新
func (r *accountRepository) Update(ctx context.Context, account *domain.Account) error {
	query := `
		UPDATE accounts
		SET email = :email, name = :name, password_hash = :password_hash, updated_at = :updated_at
		WHERE id = :id
	`

	account.UpdatedAt = time.Now()
	dbAccount := fromDomainAccount(account)

	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.NamedExecContext(ctx, query, dbAccount)
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
func (r *accountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = ?`

	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.ExecContext(ctx, query, id.String())
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

	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
