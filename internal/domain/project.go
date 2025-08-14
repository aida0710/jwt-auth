package domain

import (
	"time"

	"github.com/google/uuid"
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
	ID          uuid.UUID     `db:"id" json:"id"`
	AccountID   uuid.UUID     `db:"account_id" json:"account_id"`
	Name        string        `db:"name" json:"name"`
	Description string        `db:"description" json:"description"`
	Status      ProjectStatus `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time     `db:"updated_at" json:"updated_at"`
}

// NewProject 新しいProjectを作成
func NewProject(accountID uuid.UUID, name, description string) *Project {
	return &Project{
		ID:          uuid.Must(uuid.NewV7()), // UUID v7を使用
		AccountID:   accountID,
		Name:        name,
		Description: description,
		Status:      ProjectStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Validate プロジェクトエンティティを検証
func (p *Project) Validate() error {
	if p.AccountID == uuid.Nil {
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
