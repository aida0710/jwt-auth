package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// refreshTokenDB データベース用のリフレッシュトークン構造体
type refreshTokenDB struct {
	ID        string     `db:"id"`
	AccountID string     `db:"account_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UsedAt    *time.Time `db:"used_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *string    `db:"ip_address"`
}

// toDomain DB構造体からドメインモデルへ変換
func (r *refreshTokenDB) toDomain() (*domain.RefreshToken, error) {
	id, err := uuid.Parse(r.ID)
	if err != nil {
		return nil, err
	}
	accountID, err := uuid.Parse(r.AccountID)
	if err != nil {
		return nil, err
	}

	return &domain.RefreshToken{
		ID:        id,
		AccountID: accountID,
		TokenHash: r.TokenHash,
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
		UsedAt:    r.UsedAt,
		RevokedAt: r.RevokedAt,
		UserAgent: r.UserAgent,
		IPAddress: r.IPAddress,
	}, nil
}

// fromDomain ドメインモデルからDB構造体へ変換
func fromDomainRefreshToken(token *domain.RefreshToken) *refreshTokenDB {
	return &refreshTokenDB{
		ID:        token.ID.String(),
		AccountID: token.AccountID.String(),
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
		UsedAt:    token.UsedAt,
		RevokedAt: token.RevokedAt,
		UserAgent: token.UserAgent,
		IPAddress: token.IPAddress,
	}
}

// RefreshTokenRepository リフレッシュトークンリポジトリの実装
type RefreshTokenRepository struct {
	db *sqlx.DB
}

// NewRefreshTokenRepository 新しいリフレッシュトークンリポジトリを作成
func NewRefreshTokenRepository(db *sqlx.DB) domain.RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create 新しいリフレッシュトークンを作成
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (
			id, account_id, token_hash, expires_at, 
			created_at, user_agent, ip_address
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	dbToken := fromDomainRefreshToken(token)
	_, err := r.db.ExecContext(ctx, query,
		dbToken.ID,
		dbToken.AccountID,
		dbToken.TokenHash,
		dbToken.ExpiresAt,
		dbToken.CreatedAt,
		dbToken.UserAgent,
		dbToken.IPAddress,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash トークンハッシュからリフレッシュトークンを取得
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	var dbToken refreshTokenDB
	query := `
		SELECT 
			id, account_id, token_hash, expires_at, created_at,
			used_at, revoked_at, user_agent, ip_address
		FROM refresh_tokens 
		WHERE token_hash = ?
	`

	err := r.db.GetContext(ctx, &dbToken, query, tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return dbToken.toDomain()
}

// MarkAsUsed トークンを使用済みとしてマーク
func (r *RefreshTokenRepository) MarkAsUsed(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE refresh_tokens 
		SET used_at = ? 
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Revoke トークンを無効化
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = ? 
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id.String())
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// RevokeByAccountID アカウントIDに紐づくすべてのトークンを無効化
func (r *RefreshTokenRepository) RevokeByAccountID(ctx context.Context, accountID uuid.UUID) error {
	query := `
		UPDATE refresh_tokens 
		SET revoked_at = ? 
		WHERE account_id = ? AND revoked_at IS NULL
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), accountID.String())
	if err != nil {
		return fmt.Errorf("failed to revoke tokens by account ID: %w", err)
	}

	return nil
}

// DeleteExpired 有効期限切れのトークンを削除
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens 
		WHERE expires_at < ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return nil
}
