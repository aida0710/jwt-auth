package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// SecurityAuditLogRepository セキュリティ監査ログリポジトリの実装
type SecurityAuditLogRepository struct {
	db *sqlx.DB
}

// NewSecurityAuditLogRepository 新しいセキュリティ監査ログリポジトリを作成
func NewSecurityAuditLogRepository(db *sqlx.DB) domain.SecurityAuditLogRepository {
	return &SecurityAuditLogRepository{db: db}
}

// Create 新しいセキュリティ監査ログを作成
func (r *SecurityAuditLogRepository) Create(ctx context.Context, log *domain.SecurityAuditLog) error {
	query := `
		INSERT INTO security_audit_logs (
			id, account_id, event_type, event_description,
			ip_address, user_agent, metadata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.AccountID,
		log.EventType,
		log.EventDescription,
		log.IPAddress,
		log.UserAgent,
		log.Metadata,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create security audit log: %w", err)
	}

	return nil
}

// GetByAccountID アカウントIDからセキュリティ監査ログを取得
func (r *SecurityAuditLogRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*domain.SecurityAuditLog, error) {
	var logs []*domain.SecurityAuditLog

	query := `
		SELECT 
			id, account_id, event_type, event_description,
			ip_address, user_agent, metadata, created_at
		FROM security_audit_logs 
		WHERE account_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.SelectContext(ctx, &logs, query, accountID, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*domain.SecurityAuditLog{}, nil
		}
		return nil, fmt.Errorf("failed to get security audit logs by account ID: %w", err)
	}

	return logs, nil
}

// GetByEventType イベントタイプからセキュリティ監査ログを取得
func (r *SecurityAuditLogRepository) GetByEventType(ctx context.Context, eventType domain.SecurityEventType, limit, offset int) ([]*domain.SecurityAuditLog, error) {
	var logs []*domain.SecurityAuditLog

	query := `
		SELECT 
			id, account_id, event_type, event_description,
			ip_address, user_agent, metadata, created_at
		FROM security_audit_logs 
		WHERE event_type = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	err := r.db.SelectContext(ctx, &logs, query, eventType, limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*domain.SecurityAuditLog{}, nil
		}
		return nil, fmt.Errorf("failed to get security audit logs by event type: %w", err)
	}

	return logs, nil
}

// CountByAccountID アカウントIDごとのログ数を取得
func (r *SecurityAuditLogRepository) CountByAccountID(ctx context.Context, accountID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM security_audit_logs WHERE account_id = ?`

	err := r.db.GetContext(ctx, &count, query, accountID)
	if err != nil {
		return 0, fmt.Errorf("failed to count security audit logs: %w", err)
	}

	return count, nil
}
