package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/aida0710/jwt-auth/internal/auth"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

// AuthConfig 認証ミドルウェアの設定を保持します
type AuthConfig struct {
	JWTManager  *auth.JWTManager
	PublicPaths []string
}

// contextKey コンテキストキーの型です
type contextKey string

const (
	// AccountIDKey コンテキストからアカウントIDを取得するためのキー
	AccountIDKey contextKey = "account_id"
	// EmailKey コンテキストからメールアドレスを取得するためのキー
	EmailKey contextKey = "email"
)

// NewAuthMiddleware 認証ミドルウェアを作成
func NewAuthMiddleware(config AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// パスが認証不要リストに含まれているかチェック
			path := c.Path()
			for _, publicPath := range config.PublicPaths {
				if isPublicPath(path, publicPath) {
					return next(c)
				}
			}

			// Authorizationヘッダーからトークンを取得
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			// Bearer トークンの形式をチェック
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid authorization header format")
			}

			// fmt.Println(tokenParts[1])

			tokenString := tokenParts[1]

			// トークンを検証
			claims, err := config.JWTManager.ValidateAccessToken(tokenString)
			if err != nil {
				logSuspiciousTokenAttempt(err, c.RealIP(), c.Request().UserAgent())

				// エラーメッセージを適切に返す
				errorMsg := "invalid or expired token"
				if strings.Contains(err.Error(), "none algorithm") {
					errorMsg = "invalid token: signature required"
				} else if strings.Contains(err.Error(), "signature verification failed") {
					errorMsg = "invalid token: signature verification failed"
				} else if strings.Contains(err.Error(), "malformed") {
					errorMsg = "invalid token: malformed token"
				} else if strings.Contains(err.Error(), "expired") {
					errorMsg = "token has expired"
				}
				return echo.NewHTTPError(http.StatusUnauthorized, errorMsg)
			}

			// アカウントIDとメールを共通で使えるようにコンテキストへ設定
			c.Set(string(AccountIDKey), claims.AccountID)
			c.Set(string(EmailKey), claims.Email)

			return next(c)
		}
	}
}

// isPublicPath パスが公開パスかどうかをチェック
func isPublicPath(path, publicPath string) bool {
	if path == publicPath {
		return true
	}

	// 非公開ディレクトリ
	return false
}

// logSuspiciousTokenAttempt 不審なトークン試行をログに記録
func logSuspiciousTokenAttempt(err error, ipAddress, userAgent string) {
	var eventType domain.SecurityEventType
	var description string

	if strings.Contains(err.Error(), "none algorithm") {
		eventType = domain.EventSuspiciousLogin
		description = "Attempted to use JWT with 'none' algorithm (signature bypass attempt)"
	} else if strings.Contains(err.Error(), "signature verification failed") {
		eventType = domain.EventSuspiciousLogin
		description = "JWT signature verification failed (possible token tampering)"
	} else if strings.Contains(err.Error(), "invalid signing algorithm") {
		eventType = domain.EventSuspiciousLogin
		description = fmt.Sprintf("Invalid JWT signing algorithm attempted: %v", err)
	} else if strings.Contains(err.Error(), "malformed") {
		eventType = domain.EventSuspiciousLogin
		description = "Malformed JWT token (possible attack attempt)"
	} else {
		// 通常の期限切れなどはログに記録しない
		return
	}

	log.Warnf("[SuspiciousToken] EventType: %s | Description: %s | IP: %s | UserAgent: %s\n", eventType, description, ipAddress, userAgent)
}
