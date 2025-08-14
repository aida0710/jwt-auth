package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Account アカウントエンティティ
type Account struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	Name         string    `db:"name" json:"name"`
	PasswordHash string    `db:"password_hash" json:"-"` // JSONレスポンスには含めない
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// NewAccount 新しいAccountを作成
func NewAccount(email, name, passwordHash string) *Account {
	return &Account{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// Validate アカウントエンティティを検証
func (a *Account) Validate() error {
	if a.Email == "" {
		return ErrInvalidEmail
	}
	// 簡単なチェックのため、今後修正する必要あり
	if !strings.Contains(a.Email, "@") || !strings.Contains(a.Email, ".") {
		return ErrInvalidEmail
	}
	if len(a.Email) > MaxEmailLength {
		return ErrInvalidEmail
	}
	if a.Name == "" {
		return ErrInvalidName
	}
	if len(a.Name) > MaxNameLength {
		return ErrInvalidName
	}
	return nil
}
