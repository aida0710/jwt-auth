package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapiTypes "github.com/oapi-codegen/runtime/types"
)

// ====================================
// DTO変換関数
// ====================================

// NewAPIAccountFromEntity エンティティからAPIレスポンスに変換
func NewAPIAccountFromEntity(account *domain.Account) (api.Account, error) {
	// UUIDのパース（バリデーション済みのはずだが、念のため確認）
	accountId, err := uuid.Parse(account.ID)
	if err != nil {
		return api.Account{}, fmt.Errorf("failed to parse account ID: %w", err)
	}

	return api.Account{
		Id:        accountId,
		Email:     openapiTypes.Email(account.Email),
		Name:      account.Name,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}, nil
}

// NewAccountListResponse アカウント一覧レスポンスを生成
func NewAccountListResponse(accounts []*domain.Account, total, limit, offset int) (api.AccountListResponse, error) {
	apiAccounts := make([]api.Account, len(accounts))
	for i, account := range accounts {
		apiAccount, err := NewAPIAccountFromEntity(account)
		if err != nil {
			return api.AccountListResponse{}, fmt.Errorf("failed to convert account at index %d: %w", i, err)
		}
		apiAccounts[i] = apiAccount
	}

	return api.AccountListResponse{
		Accounts: apiAccounts,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// ====================================
// アカウント関連のハンドラー実装
// ====================================

// ListAccounts アカウント一覧を取得
func (s *Server) ListAccounts(ctx echo.Context, params api.ListAccountsParams) error {
	reqCtx := ctx.Request().Context()

	// デフォルト値の設定
	limit := 10
	offset := 0

	if params.Limit != nil && *params.Limit > 0 {
		limit = *params.Limit
	}
	if params.Offset != nil && *params.Offset >= 0 {
		offset = *params.Offset
	}

	s.logger.Info(reqCtx, "Getting accounts list",
		logger.F("limit", limit),
		logger.F("offset", offset),
	)

	accounts, total, err := s.accountUsecase.List(reqCtx, limit, offset)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get accounts", err)
		return handleAccountError(ctx, err)
	}

	response, err := NewAccountListResponse(accounts, total, limit, offset)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to convert accounts to response", err)
		return ctx.JSON(http.StatusInternalServerError, api.Error{
			Error: "Internal server error",
		})
	}
	return ctx.JSON(http.StatusOK, response)
}

// CreateAccount 新しいアカウントを作成
func (s *Server) CreateAccount(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	var req api.CreateAccountRequest
	if err := ctx.Bind(&req); err != nil {
		s.logger.Warn(reqCtx, "Invalid request body", logger.F("error", err.Error()))
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: "Invalid request body",
		})
	}

	s.logger.Info(reqCtx, "Creating new account",
		logger.F("email", req.Email),
		logger.F("name", req.Name),
	)

	input := usecase.CreateInput{
		Email: string(req.Email),
		Name:  req.Name,
	}

	account, err := s.accountUsecase.Create(reqCtx, input)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to create account", err,
			logger.F("email", req.Email),
		)
		return handleAccountError(ctx, err)
	}

	s.logger.Info(reqCtx, "Account created successfully",
		logger.F("account_id", account.ID),
	)

	apiAccount, err := NewAPIAccountFromEntity(account)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to convert account to response", err,
			logger.F("account_id", account.ID),
		)
		return ctx.JSON(http.StatusInternalServerError, api.Error{
			Error: "Internal server error",
		})
	}
	return ctx.JSON(http.StatusCreated, apiAccount)
}

// GetAccount IDでアカウントを取得
func (s *Server) GetAccount(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Getting account by ID",
		logger.F("account_id", accountId),
	)

	account, err := s.accountUsecase.GetByID(reqCtx, accountId.String())
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get account", err,
			logger.F("account_id", accountId),
		)
		return handleAccountError(ctx, err)
	}

	apiAccount, err := NewAPIAccountFromEntity(account)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to convert account to response", err,
			logger.F("account_id", account.ID),
		)
		return ctx.JSON(http.StatusInternalServerError, api.Error{
			Error: "Internal server error",
		})
	}
	return ctx.JSON(http.StatusOK, apiAccount)
}

// UpdateAccount アカウントを更新
func (s *Server) UpdateAccount(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	var req api.UpdateAccountRequest
	if err := ctx.Bind(&req); err != nil {
		s.logger.Warn(reqCtx, "Invalid request body", logger.F("error", err.Error()))
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: "Invalid request body",
		})
	}

	s.logger.Info(reqCtx, "Updating account",
		logger.F("account_id", accountId),
	)

	input := usecase.UpdateInput{}
	if req.Email != nil {
		email := string(*req.Email)
		input.Email = &email
	}
	if req.Name != nil {
		input.Name = req.Name
	}

	account, err := s.accountUsecase.Update(reqCtx, accountId.String(), input)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to update account", err,
			logger.F("account_id", accountId),
		)
		return handleAccountError(ctx, err)
	}

	s.logger.Info(reqCtx, "Account updated successfully",
		logger.F("account_id", accountId),
	)

	apiAccount, err := NewAPIAccountFromEntity(account)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to convert account to response", err,
			logger.F("account_id", account.ID),
		)
		return ctx.JSON(http.StatusInternalServerError, api.Error{
			Error: "Internal server error",
		})
	}
	return ctx.JSON(http.StatusOK, apiAccount)
}

// DeleteAccount アカウントを削除
func (s *Server) DeleteAccount(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Deleting account",
		logger.F("account_id", accountId),
	)

	err := s.accountUsecase.Delete(reqCtx, accountId.String())
	if err != nil {
		s.logger.Error(reqCtx, "Failed to delete account", err,
			logger.F("account_id", accountId),
		)
		return handleAccountError(ctx, err)
	}

	s.logger.Info(reqCtx, "Account deleted successfully",
		logger.F("account_id", accountId),
	)

	return ctx.NoContent(http.StatusNoContent)
}

// ====================================
// エラーハンドリング
// ====================================

// handleAccountError アカウント関連のエラーをHTTPレスポンスに変換
func handleAccountError(ctx echo.Context, err error) error {
	// エラーマッピングから適切なステータスコードを探す
	if errors.Is(err, domain.ErrAccountNotFound) {
		return ctx.JSON(http.StatusNotFound, api.Error{
			Error: err.Error(),
		})
	}
	if errors.Is(err, domain.ErrDuplicateEmail) {
		return ctx.JSON(http.StatusConflict, api.Error{
			Error: err.Error(),
		})
	}
	if errors.Is(err, domain.ErrInvalidEmail) || errors.Is(err, domain.ErrInvalidName) ||
		errors.Is(err, domain.ErrInvalidID) || errors.Is(err, domain.ErrInvalidAccountID) {
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: err.Error(),
		})
	}

	// デフォルトのエラーレスポンス
	return ctx.JSON(http.StatusInternalServerError, api.Error{
		Error: "Internal server error",
	})
}
