package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aida0710/jwt-auth/internal/auth"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

// AuthUsecase 認証関連のユースケース
type AuthUsecase struct {
	accountRepo       domain.AccountRepository
	refreshTokenRepo  domain.RefreshTokenRepository
	securityAuditRepo domain.SecurityAuditLogRepository
	jwtManager        *auth.JWTManager
}

// NewAuthUsecase 新しい認証ユースケースを作成
func NewAuthUsecase(
	accountRepo domain.AccountRepository,
	refreshTokenRepo domain.RefreshTokenRepository,
	securityAuditRepo domain.SecurityAuditLogRepository,
	jwtManager *auth.JWTManager,
) *AuthUsecase {
	return &AuthUsecase{
		accountRepo:       accountRepo,
		refreshTokenRepo:  refreshTokenRepo,
		securityAuditRepo: securityAuditRepo,
		jwtManager:        jwtManager,
	}
}

// SignUpInput サインアップの入力
type SignUpInput struct {
	Email    string
	Password string
	Name     string
}

// LoginInput ログインの入力
type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
	IPAddress string
}

// AuthTokens 認証トークンのペア
type AuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	Account      *domain.Account
}

// SignUp 新規アカウントを作成
func (u *AuthUsecase) SignUp(ctx context.Context, input SignUpInput) (*AuthTokens, error) {
	existing, err := u.accountRepo.GetByEmail(ctx, input.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, fmt.Errorf("failed to check existing account: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	passwordHash, err := auth.HashPassword(input.Password)
	fmt.Printf("passwordHash: %s\n", passwordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// NewAccountを使用してUUID v4で作成
	account := domain.NewAccount(input.Email, input.Name, passwordHash)

	// アカウントを検証
	if err := account.Validate(); err != nil {
		return nil, err
	}

	// データベースに保存
	if err := u.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// トークンを生成
	return u.generateTokens(ctx, account, "", "")
}

// Login メールとパスワードでログイン
func (u *AuthUsecase) Login(ctx context.Context, input LoginInput) (*AuthTokens, error) {
	// アカウントを取得
	account, err := u.accountRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	if err := auth.VerifyPassword(input.Password, account.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// トークンを生成
	return u.generateTokens(ctx, account, input.UserAgent, input.IPAddress)
}

// RefreshToken リフレッシュトークンを使用して新しいトークンを生成
func (u *AuthUsecase) RefreshToken(ctx context.Context, refreshToken string, userAgent, ipAddress string) (*AuthTokens, error) {
	// リフレッシュトークンを検証
	claims, err := u.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		// 不正なトークンの試行をログに記録
		if strings.Contains(err.Error(), "none algorithm") ||
			strings.Contains(err.Error(), "signature verification failed") ||
			strings.Contains(err.Error(), "invalid signing algorithm") {
			u.logSecurityEvent(ctx, uuid.Nil,
				domain.EventSuspiciousLogin,
				fmt.Sprintf("Invalid refresh token attempt: %v", err),
				userAgent, ipAddress)
		}
		return nil, domain.ErrInvalidToken
	}

	// データベースからトークンを取得
	tokenHash := auth.HashToken(refreshToken)
	storedToken, err := u.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	// 使用済みトークンの再利用を検出（セキュリティ侵害の可能性）
	if storedToken.UsedAt != nil {
		// セキュリティ侵害の可能性があるため、このアカウントのすべてのリフレッシュトークンを無効化
		if err := u.refreshTokenRepo.RevokeByAccountID(ctx, storedToken.AccountID); err != nil {
			// エラーでも続行（セキュリティを優先）
			fmt.Printf("Failed to revoke tokens for account %s: %v\n", storedToken.AccountID, err)
		}

		// セキュリティイベントを記録
		u.logSecurityEvent(ctx, storedToken.AccountID,
			domain.EventTokenReuseDetected,
			"Attempted reuse of used refresh token detected. All tokens have been revoked for security.",
			userAgent, ipAddress)

		return nil, domain.ErrTokenCompromised
	}

	// トークンの有効性を確認（有効期限切れ、無効化済み）
	if !storedToken.IsValid() {
		return nil, domain.ErrInvalidToken
	}

	// claims.AccountIDをUUIDに変換
	accountID, err := uuid.Parse(claims.AccountID)
	if err != nil {
		return nil, fmt.Errorf("invalid account ID in token: %w", err)
	}

	// アカウントを取得
	account, err := u.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidToken
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	// 古いトークンを使用済みにマーク
	if err := u.refreshTokenRepo.MarkAsUsed(ctx, storedToken.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// 新しいトークンを生成
	return u.generateTokens(ctx, account, userAgent, ipAddress)
}

// Logout リフレッシュトークンを無効化
func (u *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	// トークンハッシュを計算
	tokenHash := auth.HashToken(refreshToken)

	// トークンを取得
	storedToken, err := u.refreshTokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// トークンが見つからない場合も正常終了
			return nil
		}
		return fmt.Errorf("failed to get refresh token: %w", err)
	}

	// トークンを無効化
	if err := u.refreshTokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// LogoutAll アカウントのすべてのリフレッシュトークンを無効化
func (u *AuthUsecase) LogoutAll(ctx context.Context, accountID uuid.UUID) error {
	if err := u.refreshTokenRepo.RevokeByAccountID(ctx, accountID); err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}
	return nil
}

// logSecurityEvent セキュリティイベントをログに記録
func (u *AuthUsecase) logSecurityEvent(
	ctx context.Context,
	accountID uuid.UUID,
	eventType domain.SecurityEventType,
	description string,
	userAgent, ipAddress string,
) {
	// セキュリティ監査ログを作成
	var userAgentPtr, ipAddressPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	auditLog, err := domain.NewSecurityAuditLog(
		accountID,
		eventType,
		description,
		ipAddressPtr,
		userAgentPtr,
		nil, // 追加メタデータがあればここに設定
	)
	if err != nil {
		fmt.Printf("[ERROR] Failed to create security audit log: %v\n", err)
		return
	}

	if u.securityAuditRepo != nil {
		if err := u.securityAuditRepo.Create(ctx, auditLog); err != nil {
			fmt.Printf("[ERROR] Failed to save security audit log: %v\n", err)
		}
	}

	log.Warnf("[SECURITY ALERT] AccountID: %s, Event: %s, Description: %s, IP: %s\n", accountID.String(), eventType, description, ipAddress)
}

// generateTokens アクセストークンとリフレッシュトークンを生成
func (u *AuthUsecase) generateTokens(ctx context.Context, account *domain.Account, userAgent, ipAddress string) (*AuthTokens, error) {
	// アクセストークンを生成
	accessToken, err := u.jwtManager.GenerateAccessToken(account.ID, account.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// リフレッシュトークンを生成
	refreshToken, tokenID, err := u.jwtManager.GenerateRefreshToken(account.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// リフレッシュトークンをデータベースに保存
	var userAgentPtr, ipAddressPtr *string
	if userAgent != "" {
		userAgentPtr = &userAgent
	}
	if ipAddress != "" {
		ipAddressPtr = &ipAddress
	}

	storedToken := domain.NewRefreshToken(
		account.ID,
		auth.HashToken(refreshToken),
		time.Now().Add(30*24*time.Hour), // 30日
		userAgentPtr,
		ipAddressPtr,
	)
	storedToken.ID = tokenID // JWTから生成されたtokenIDを使用

	if err := u.refreshTokenRepo.Create(ctx, storedToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// パスワードハッシュを除外したアカウント情報を返す
	accountCopy := *account
	accountCopy.PasswordHash = ""

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1時間（秒）
		Account:      &accountCopy,
	}, nil
}
