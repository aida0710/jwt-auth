package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword パスワードをハッシュ化します
func HashPassword(password string) (string, error) {
	// bcrypt costは通常10〜12の範囲で設定するらしい。
	// 以下のサイトに仕組みが簡単に記載されていた。
	// https://qiita.com/iheuko/items/e1be4b646be11e329cd8
	//hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword パスワードとハッシュを検証します
func VerifyPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
