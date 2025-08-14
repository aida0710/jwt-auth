package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/labstack/echo/v4"
	openapiTypes "github.com/oapi-codegen/runtime/types"
)

// AuthHandler 認証関連のハンドラー
type AuthHandler struct {
	authUsecase *usecase.AuthUsecase
}

// NewAuthHandler 新しい認証ハンドラーを作成
func NewAuthHandler(authUsecase *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{
		authUsecase: authUsecase,
	}
}

// SignUp 新規アカウント登録
func (h *AuthHandler) SignUp(c echo.Context) error {
	var req api.SignUpRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email, password and name are required")
	}

	if len(req.Password) < 8 {
		return echo.NewHTTPError(http.StatusBadRequest, "password must be at least 8 characters")
	}

	if len(req.Password) > 60 {
		// bcryptは最大72バイト (ASCII文字なら72文字) まで
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("password must be less than 60 characters"))
	}

	tokens, err := h.authUsecase.SignUp(c.Request().Context(), usecase.SignUpInput{
		Email:    string(req.Email),
		Password: req.Password,
		Name:     req.Name,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrEmailAlreadyExists), errors.Is(err, domain.ErrDuplicateEmail):
			return echo.NewHTTPError(http.StatusConflict, "email already exists")
		case errors.Is(err, domain.ErrInvalidEmail):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid email address")
		case errors.Is(err, domain.ErrInvalidName):
			return echo.NewHTTPError(http.StatusBadRequest, "invalid name")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create account")
		}
	}

	return c.JSON(http.StatusCreated, api.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		Account: api.Account{
			Id:        tokens.Account.ID,
			Email:     openapiTypes.Email(tokens.Account.Email),
			Name:      tokens.Account.Name,
			CreatedAt: tokens.Account.CreatedAt,
			UpdatedAt: tokens.Account.UpdatedAt,
		},
	})
}

// Login メールとパスワードでログイン
func (h *AuthHandler) Login(c echo.Context) error {
	var req api.LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "email and password are required")
	}

	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()

	tokens, err := h.authUsecase.Login(c.Request().Context(), usecase.LoginInput{
		Email:     string(req.Email),
		Password:  req.Password,
		UserAgent: userAgent,
		IPAddress: ipAddress,
	})

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to login")
		}
	}

	return c.JSON(http.StatusOK, api.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		Account: api.Account{
			Id:        tokens.Account.ID,
			Email:     openapiTypes.Email(tokens.Account.Email),
			Name:      tokens.Account.Name,
			CreatedAt: tokens.Account.CreatedAt,
			UpdatedAt: tokens.Account.UpdatedAt,
		},
	})
}

// RefreshToken リフレッシュトークンを使用して新しいトークンを生成
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req api.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	userAgent := c.Request().UserAgent()
	ipAddress := c.RealIP()

	tokens, err := h.authUsecase.RefreshToken(
		c.Request().Context(),
		req.RefreshToken,
		userAgent,
		ipAddress,
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTokenCompromised):
			// セキュリティ侵害の可能性がある場合は、明確にユーザーに通知
			return echo.NewHTTPError(http.StatusUnauthorized, "Security alert: This refresh token has already been used. For your security, all tokens have been revoked. Please login again.")
		case errors.Is(err, domain.ErrInvalidToken), errors.Is(err, domain.ErrTokenExpired):
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired refresh token")
		default:
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to refresh token")
		}
	}

	return c.JSON(http.StatusOK, api.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokens.ExpiresIn,
		Account: api.Account{
			Id:        tokens.Account.ID,
			Email:     openapiTypes.Email(tokens.Account.Email),
			Name:      tokens.Account.Name,
			CreatedAt: tokens.Account.CreatedAt,
			UpdatedAt: tokens.Account.UpdatedAt,
		},
	})
}

// Logout リフレッシュトークンを無効化
func (h *AuthHandler) Logout(c echo.Context) error {
	var req api.LogoutRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh_token is required")
	}

	if err := h.authUsecase.Logout(c.Request().Context(), req.RefreshToken); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to logout")
	}

	// 204 No Content を返す
	return c.NoContent(http.StatusNoContent)
}
