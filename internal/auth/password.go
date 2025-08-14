package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword パスワードをハッシュ化します
func HashPassword(password string) (string, error) {
	// bcrypt cost は
	//hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword パスワードとハッシュを検証します
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
