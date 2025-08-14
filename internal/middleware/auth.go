package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/aida0710/jwt-auth/internal/auth"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/labstack/echo/v4"
)

// AuthConfig 認証ミドルウェアの設定を保持します
type AuthConfig struct {
	JWTManager *auth.JWTManager
	// 認証不要のパスのリスト
	PublicPaths []string
	// セキュリティ監査ログリポジトリ（オプション）
	SecurityAuditRepo domain.SecurityAuditLogRepository
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

			fmt.Println(tokenParts[1])

			tokenString := tokenParts[1]

			// トークンを検証
			claims, err := config.JWTManager.ValidateAccessToken(tokenString)
			if err != nil {
				// セキュリティ上注意すべきエラーを記録
				if config.SecurityAuditRepo != nil {
					logSuspiciousTokenAttempt(c.Request().Context(), config.SecurityAuditRepo, err, c.RealIP(), c.Request().UserAgent())
				}

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

			// アカウントIDとメールをコンテキストに設定
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

	// プレフィックスマッチ（ワイルドカード対応）
	if strings.HasSuffix(publicPath, "*") {
		prefix := strings.TrimSuffix(publicPath, "*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// logSuspiciousTokenAttempt 不審なトークン試行をログに記録
func logSuspiciousTokenAttempt(ctx context.Context, repo domain.SecurityAuditLogRepository, err error, ipAddress, userAgent string) {
	// 特に重要なセキュリティイベントを判定
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

	// セキュリティ監査ログを作成
	var ipAddressPtr, userAgentPtr *string
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}
	if userAgent != "" {
		userAgentPtr = &userAgent
	}

	// accountIDがない場合は"UNKNOWN"を使用
	auditLog, err := domain.NewSecurityAuditLog(
		"UNKNOWN", // トークンが無効なためアカウントIDが不明
		eventType,
		description,
		ipAddressPtr,
		userAgentPtr,
		nil,
	)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create security audit log: %v\n", err)
		return
	}

	// データベースに記録（エラーはログに出力するが処理は継続）
	if err := repo.Create(ctx, auditLog); err != nil {
		fmt.Printf("[ERROR] Failed to save security audit log: %v\n", err)
	}

	// コンソールにも警告を出力
	fmt.Printf("[SECURITY WARNING] %s from IP: %s, UserAgent: %s\n", description, ipAddress, userAgent)
}
