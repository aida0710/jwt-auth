package domain

import "github.com/google/uuid"

// IsValidUUID 文字列が有効なUUIDか確認
func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// ValidateID UUIDフォーマットを検証して適切なエラーを返す
func ValidateID(id string, errType error) error {
	if id == "" {
		return errType
	}
	if !IsValidUUID(id) {
		return errType
	}
	return nil
}

// ValidateAccountID アカウントIDを検証
func ValidateAccountID(id string) error {
	return ValidateID(id, ErrInvalidAccountID)
}

// ValidateProjectID プロジェクトIDを検証
func ValidateProjectID(id string) error {
	return ValidateID(id, ErrInvalidID)
}

// ビジネスルール用の定数
const (
	MaxProjectsPerAccount = 10
	MaxNameLength         = 255
	MaxEmailLength        = 255
)
