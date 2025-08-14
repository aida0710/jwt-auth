package domain

import (
	"context"

	"github.com/google/uuid"
)

// AccountRepository アカウントリポジトリのインターフェースを定義
type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	GetByID(ctx context.Context, id uuid.UUID) (*Account, error)
	GetByEmail(ctx context.Context, email string) (*Account, error)
	List(ctx context.Context, limit, offset int) ([]*Account, error)
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int, error)
}

// ProjectRepository プロジェクトリポジトリのインターフェースを定義
type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*Project, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*Project, error)
	List(ctx context.Context, limit, offset int) ([]*Project, error)
	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByAccountID(ctx context.Context, accountID uuid.UUID) error
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int, error)
	Count(ctx context.Context) (int, error)
}

// RefreshTokenRepository リフレッシュトークンリポジトリのインターフェースを定義
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	MarkAsUsed(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByAccountID(ctx context.Context, accountID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// SecurityAuditLogRepository セキュリティ監査ログリポジトリのインターフェースを定義
type SecurityAuditLogRepository interface {
	Create(ctx context.Context, log *SecurityAuditLog) error
	GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*SecurityAuditLog, error)
	GetByEventType(ctx context.Context, eventType SecurityEventType, limit, offset int) ([]*SecurityAuditLog, error)
	CountByAccountID(ctx context.Context, accountID uuid.UUID) (int, error)
}
