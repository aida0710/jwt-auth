package domain

import (
	"time"
)

// ProjectStatus プロジェクトのステータス
type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "active"
	ProjectStatusInactive ProjectStatus = "inactive"
	ProjectStatusArchived ProjectStatus = "archived"
)

// Project プロジェクトエンティティ
type Project struct {
	ID          string        `db:"id" json:"id"`
	AccountID   string        `db:"account_id" json:"account_id"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
	Status      ProjectStatus `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}

// Validate プロジェクトエンティティを検証
func (p *Project) Validate() error {
	if p.AccountID == "" {
		return ErrInvalidAccountID
	}
	if p.Name == "" {
		return ErrInvalidName
	}
	if len(p.Name) > MaxNameLength {
		return ErrInvalidName
	}
	if p.Status == "" {
		p.Status = ProjectStatusActive
	}
	if !p.IsValidStatus() {
		return ErrInvalidStatus
	}
	return nil
}

// IsValidStatus ステータスが有効か確認
func (p *Project) IsValidStatus() bool {
	switch p.Status {
	case ProjectStatusActive, ProjectStatusInactive, ProjectStatusArchived:
		return true
	default:
		return false
	}
}
