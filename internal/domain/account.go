package domain

import (
	"strings"
	"time"
)

// Account アカウントエンティティ
type Account struct {
	ID        string    `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Name      string    `db:"name" json:"name"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
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
