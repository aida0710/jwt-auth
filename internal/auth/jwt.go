package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTConfig JWT設定を保持
type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	Audience           []string
}

// Claims JWTのカスタムクレームを定義
type Claims struct {
	AccountID string `json:"account_id"` // JWTペイロードは文字列
	Email     string `json:"email"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims リフレッシュトークンのクレームを定義
type RefreshTokenClaims struct {
	TokenID   string `json:"token_id"`   // JWTペイロードは文字列
	AccountID string `json:"account_id"` // JWTペイロードは文字列
	jwt.RegisteredClaims
}

// JWTManager JWTトークンの管理
type JWTManager struct {
	config JWTConfig
}

// NewJWTManager 新しいJWTManagerを作成
func NewJWTManager(config JWTConfig) *JWTManager {
	// デフォルト値を設定
	if config.AccessTokenExpiry == 0 {
		config.AccessTokenExpiry = time.Hour
	}
	if config.RefreshTokenExpiry == 0 {
		config.RefreshTokenExpiry = time.Hour * 24 * 30
	}

	return &JWTManager{
		config: config,
	}
}

// GenerateAccessToken アクセストークンを生成
func (m *JWTManager) GenerateAccessToken(accountID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		AccountID: accountID.String(), // UUID→文字列変換
		Email:     email,
		RegisteredClaims: jwt.RegisteredClaims{
			// トークンの有効期限を設定（Missing Expiration Vulnerabilityを防ぐ）
			// 参照: https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Subject:   accountID.String(),
			ID:        uuid.Must(uuid.NewV7()).String(), // UUID v7を使用
			Audience:  m.config.Audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.config.AccessTokenSecret))
}

// GenerateRefreshToken リフレッシュトークンを生成
func (m *JWTManager) GenerateRefreshToken(accountID uuid.UUID) (string, uuid.UUID, error) {
	// リフレッシュトークン用のユニークIDを生成（UUID v7）
	tokenID := uuid.Must(uuid.NewV7())

	now := time.Now()
	claims := &RefreshTokenClaims{
		TokenID:   tokenID.String(),   // UUID→文字列変換
		AccountID: accountID.String(), // UUID→文字列変換
		RegisteredClaims: jwt.RegisteredClaims{
			// トークンの有効期限を設定（Missing Expiration Vulnerabilityを防ぐ）
			// 参照: https://auth0.com/blog/a-look-at-the-latest-draft-for-jwt-bcp/
			ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.config.Issuer,
			Subject:   accountID.String(),
			ID:        tokenID.String(),
			Audience:  m.config.Audience,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.RefreshTokenSecret)) // ここで署名
	if err != nil {
		return "", uuid.Nil, err
	}

	return tokenString, tokenID, nil // tokenIDはUUIDで返す
}

// validateToken 汎用的なトークン検証
func (m *JWTManager) validateToken(tokenString string, claims jwt.Claims, secret []byte, tokenType string) error {
	// トークンの基本的な構造をチェック（3つのパートがあるか）
	// Malformed Token Attack / Token Manipulation Attackを防ぐ
	// 参照: https://portswigger.net/web-security/jwt
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return fmt.Errorf("invalid %s structure: expected 3 parts, got %d", tokenType, len(parts))
	}

	// 各パートが空でないことを確認
	// Signature Stripping Attack（署名部分を削除する攻撃）を防ぐ
	// 参照: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	for i, part := range parts {
		if part == "" {
			return fmt.Errorf("invalid %s: part %d is empty", tokenType, i+1)
		}
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// アルゴリズムを厳密にチェック（HS256のみ許可）
		// Algorithm Confusion Attack（RS256をHS256に偽装する攻撃）を防ぐ
		// 参照: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		// 参照: https://portswigger.net/web-security/jwt/algorithm-confusion
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("invalid signing algorithm: %v (expected HS256)", token.Header["alg"])
		}

		// Noneアルゴリズムを明示的に拒否
		// None Algorithm Attack（署名検証をバイパスする攻撃）を防ぐ
		// 参照: https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
		// 参照: https://portswigger.net/web-security/jwt#accepting-tokens-with-no-signature
		if token.Method.Alg() == "none" || token.Method.Alg() == "" {
			return nil, fmt.Errorf("none algorithm is not allowed")
		}

		// 署名方法の型をチェック
		// Algorithm Substitution Attack（異なる署名アルゴリズムへの置換）を防ぐ
		// 参照: https://www.rfc-editor.org/rfc/rfc8725#section-3.1
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method type: %T", token.Method)
		}

		return secret, nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenMalformed):
			return fmt.Errorf("%s is malformed", tokenType)
		case errors.Is(err, jwt.ErrTokenExpired):
			// Token Replay Attack（期限切れトークンの再利用）を防ぐ
			// 参照: https://datatracker.ietf.org/doc/html/rfc8725#section-3.10
			return fmt.Errorf("%s has expired", tokenType)
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			// Clock Skew Attack（時刻のずれを悪用した攻撃）への対処
			// 参照: https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.5
			return fmt.Errorf("%s is not valid yet", tokenType)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			// Signature Validation Bypass Attackを防ぐ
			// 参照: https://portswigger.net/web-security/jwt#jwt-signature-verification
			return fmt.Errorf("%s signature verification failed", tokenType)
		default:
			return fmt.Errorf("%s validation failed: %v", tokenType, err)
		}
	}

	// トークンが有効であることを確認
	// 最終的な整合性チェック
	if !token.Valid {
		return fmt.Errorf("%s is invalid", tokenType)
	}

	return nil
}

// validateStandardClaims 標準的なクレームの検証
func (m *JWTManager) validateStandardClaims(issuer string, audience []string) error {
	// Issuerの検証
	// Token Substitution Attack（異なる発行者のトークンを使用する攻撃）を防ぐ
	// 参照: https://datatracker.ietf.org/doc/html/rfc8725#section-3.5
	if issuer != m.config.Issuer {
		return fmt.Errorf("invalid issuer: expected %s, got %s", m.config.Issuer, issuer)
	}

	// Audienceの検証
	// Token Confusion Attack（異なる対象者向けのトークンを誤用する攻撃）を防ぐ
	// 参照: https://datatracker.ietf.org/doc/html/rfc8725#section-3.9
	// 参照: https://www.rfc-editor.org/rfc/rfc7519#section-4.1.3
	/*if len(m.config.Audience) > 0 {
		validAudience := false
		for _, tokenAud := range audience {
			if slices.Contains(m.config.Audience, tokenAud) {
				validAudience = true //一つでも見つかればいいらしい。rfc7519
				break
			}
		}
		if !validAudience {
			return fmt.Errorf("invalid audience: token audience %v does not match expected %v", audience, m.config.Audience)
		}
	}*/

	// rfcの推奨ではないが、完全一致のほうが堅牢なので完全一致で実装。
	// マイクロサービスで同一のシークレットを使用する場合、Audienceの完全一致を要求することで、トークンの誤用を防げるかな？
	if len(m.config.Audience) > 0 {
		if !audienceExactMatch(audience, m.config.Audience) {
			return fmt.Errorf("audience mismatch: token has %v, expected exactly %v",
				audience, m.config.Audience)
		}
	}

	return nil
}

// ValidateAccessToken アクセストークンを検証
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	// 共通のトークン検証
	if err := m.validateToken(tokenString, claims, []byte(m.config.AccessTokenSecret), "token"); err != nil {
		return nil, err
	}

	// 必須フィールドの存在確認
	// Claim Tampering Attack（クレームの改ざん）を防ぐ
	// 参照: https://auth0.com/docs/ja-jp/secure/tokens/json-web-tokens/json-web-token-claims
	if claims.AccountID == "" {
		return nil, fmt.Errorf("missing account ID in claims")
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("missing email in claims")
	}

	// 標準クレームの検証
	if err := m.validateStandardClaims(claims.Issuer, claims.Audience); err != nil {
		return nil, err
	}

	// AccountIDがUUID形式であることを検証
	if _, err := uuid.Parse(claims.AccountID); err != nil {
		return nil, fmt.Errorf("invalid account ID format: %v", err)
	}

	return claims, nil
}

// audienceExactMatch 2つのaudienceスライスが完全一致するか確認
func audienceExactMatch(tokenAud, configAud []string) bool {
	if len(tokenAud) != len(configAud) {
		return false
	}

	// 両方をソート
	sortedToken := make([]string, len(tokenAud))
	sortedConfig := make([]string, len(configAud))
	copy(sortedToken, tokenAud)
	copy(sortedConfig, configAud)
	sort.Strings(sortedToken)
	sort.Strings(sortedConfig)

	// 要素ごとに比較
	for i := range sortedToken {
		if sortedToken[i] != sortedConfig[i] {
			return false
		}
	}

	return true
}

// ValidateRefreshToken はリフレッシュトークンを検証します
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	claims := &RefreshTokenClaims{}

	// 共通のトークン検証
	if err := m.validateToken(tokenString, claims, []byte(m.config.RefreshTokenSecret), "refresh token"); err != nil {
		return nil, err
	}

	// 必須フィールドの存在確認
	// Claim Tampering Attack（クレームの改ざん）を防ぐ
	// Token ID は Token Replay Attack（トークンの再利用攻撃）を防ぐためにも重要
	// 参照: https://auth0.com/docs/secure/tokens/refresh-tokens/refresh-token-rotation
	if claims.TokenID == "" {
		return nil, fmt.Errorf("missing token ID in refresh token claims")
	}
	if claims.AccountID == "" {
		return nil, fmt.Errorf("missing account ID in refresh token claims")
	}

	// TokenIDがUUID形式であることを検証
	if _, err := uuid.Parse(claims.TokenID); err != nil {
		return nil, fmt.Errorf("invalid token ID format: %v", err)
	}
	// AccountIDがUUID形式であることを検証
	if _, err := uuid.Parse(claims.AccountID); err != nil {
		return nil, fmt.Errorf("invalid account ID format: %v", err)
	}

	// 標準クレームの検証
	if err := m.validateStandardClaims(claims.Issuer, claims.Audience); err != nil {
		return nil, err
	}

	return claims, nil
}

// HashToken はトークンをハッシュ化します(トークンのハッシュ化なのでソルトはかけない。
// Token Storage Security - 平文でのトークン保存を避ける
// 参照: https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html#token-storage-on-server-side
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateSecureToken はセキュアなランダムトークンを生成します
// Weak Random Generation Vulnerabilityを防ぐ
// 参照: https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html#secure-random-number-generation
//
// 使用例:
// - パスワードリセットトークン
// - メール確認トークン
// - APIキー生成
// - セッションID生成
//
// TODO: 現在未使用だが、今後の機能拡張（パスワードリセット、メール確認など）で使用予定
func GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
