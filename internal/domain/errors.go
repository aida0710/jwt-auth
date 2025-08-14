package domain

import (
	"errors"
	"fmt"
)

var (
	ErrAccountNotFound    = errors.New("account not found")
	ErrInvalidEmail       = errors.New("invalid email address")
	ErrInvalidName        = errors.New("invalid name")
	ErrDuplicateEmail     = errors.New("email already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")

	ErrProjectNotFound      = errors.New("project not found")
	ErrInvalidAccountID     = errors.New("invalid account id")
	ErrInvalidStatus        = errors.New("invalid project status")
	ErrProjectLimitExceeded = errors.New("project limit exceeded (max: 10)")

	ErrInvalidID = errors.New("invalid id format")
	ErrNotFound  = errors.New("not found")

	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidToken       = errors.New("invalid or expired token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenCompromised   = errors.New("token may be compromised - all tokens have been revoked for security")
	ErrUnauthorized       = errors.New("unauthorized")
)

// ValidationError バリデーションエラーを表す構造体
type ValidationError struct {
	Field   string
	Message string
}

// Error errorインターフェースを実装
func (v ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", v.Field, v.Message)
}

// ValidationErrors 複数のバリデーションエラーを保持
type ValidationErrors struct {
	Errors []ValidationError
}

// Error errorインターフェースを実装
func (v *ValidationErrors) Error() string {
	if v == nil || len(v.Errors) == 0 {
		return "validation error"
	}
	return fmt.Sprintf("validation error: %s - %s", v.Errors[0].Field, v.Errors[0].Message)
}
