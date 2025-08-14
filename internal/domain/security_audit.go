package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SecurityEventType セキュリティイベントの種類
type SecurityEventType string

const (
	// EventTokenReuseDetected 使用済みトークンの再利用検出
	EventTokenReuseDetected SecurityEventType = "TOKEN_REUSE_DETECTED"
	// EventAllTokensRevoked すべてのトークンを無効化
	EventAllTokensRevoked SecurityEventType = "ALL_TOKENS_REVOKED"
	// EventSuspiciousLogin 疑わしいログイン試行
	EventSuspiciousLogin SecurityEventType = "SUSPICIOUS_LOGIN"
	// EventPasswordChanged パスワード変更
	EventPasswordChanged SecurityEventType = "PASSWORD_CHANGED"
	// EventAccountLocked アカウントロック
	EventAccountLocked SecurityEventType = "ACCOUNT_LOCKED"
	// EventMultipleFailedLogins 複数回のログイン失敗
	EventMultipleFailedLogins SecurityEventType = "MULTIPLE_FAILED_LOGINS"
)

// SecurityAuditLog セキュリティ監査ログのドメインモデル
type SecurityAuditLog struct {
	ID               uuid.UUID         `db:"id"`
	AccountID        uuid.UUID         `db:"account_id"`
	EventType        SecurityEventType `db:"event_type"`
	EventDescription string            `db:"event_description"`
	IPAddress        *string           `db:"ip_address"`
	UserAgent        *string           `db:"user_agent"`
	Metadata         json.RawMessage   `db:"metadata"`
	CreatedAt        time.Time         `db:"created_at"`
}

// SecurityAuditMetadata メタデータの構造
type SecurityAuditMetadata map[string]interface{}

// NewSecurityAuditLog 新しいセキュリティ監査ログを作成
func NewSecurityAuditLog(
	accountID uuid.UUID,
	eventType SecurityEventType,
	description string,
	ipAddress, userAgent *string,
	metadata SecurityAuditMetadata,
) (*SecurityAuditLog, error) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err != nil {
			return nil, err
		}
		metadataJSON = data
	}

	return &SecurityAuditLog{
		ID:               uuid.New(),
		AccountID:        accountID,
		EventType:        eventType,
		EventDescription: description,
		IPAddress:        ipAddress,
		UserAgent:        userAgent,
		Metadata:         metadataJSON,
		CreatedAt:        time.Now(),
	}, nil
}
