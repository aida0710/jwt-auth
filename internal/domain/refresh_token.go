package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken リフレッシュトークンのドメインモデル
type RefreshToken struct {
	ID        uuid.UUID  `db:"id"`
	AccountID uuid.UUID  `db:"account_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	CreatedAt time.Time  `db:"created_at"`
	UsedAt    *time.Time `db:"used_at"`
	RevokedAt *time.Time `db:"revoked_at"`
	UserAgent *string    `db:"user_agent"`
	IPAddress *string    `db:"ip_address"`
}

// NewRefreshToken 新しいRefreshTokenを作成（UUID v7を使用）
func NewRefreshToken(accountID uuid.UUID, tokenHash string, expiresAt time.Time, userAgent, ipAddress *string) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.Must(uuid.NewV7()), // UUID v7を使用
		AccountID: accountID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}
}

// IsValid トークンが有効かどうかを確認します
func (rt *RefreshToken) IsValid() bool {
	now := time.Now()
	// 有効期限切れ、使用済み、無効化済みでないことを確認
	return rt.ExpiresAt.After(now) && rt.UsedAt == nil && rt.RevokedAt == nil
}

// MarkAsUsed トークンを使用済みとしてマークします
func (rt *RefreshToken) MarkAsUsed() {
	now := time.Now()
	rt.UsedAt = &now
}

// Revoke トークンを無効化します
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.RevokedAt = &now
}
