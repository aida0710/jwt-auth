package handler

import (
	"errors"
	"net/http"

	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/labstack/echo/v4"
	openapiTypes "github.com/oapi-codegen/runtime/types"
)

// NewAPIAccountFromEntity エンティティからAPIレスポンスに変換
func NewAPIAccountFromEntity(account *domain.Account) api.Account {
	return api.Account{
		Id:        account.ID,
		Email:     openapiTypes.Email(account.Email),
		Name:      account.Name,
		CreatedAt: account.CreatedAt,
		UpdatedAt: account.UpdatedAt,
	}
}

// ListAccounts アカウント一覧を取得
func (s *Server) ListAccounts(ctx echo.Context) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Getting accounts list")

	// すべてのアカウントを取得
	accounts, err := s.accountUsecase.List(reqCtx)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get accounts", err)
		return handleAccountError(ctx, err)
	}

	// エンティティからAPIレスポンスに変換
	apiAccounts := make([]api.Account, len(accounts))
	for i, account := range accounts {
		apiAccounts[i] = NewAPIAccountFromEntity(account)
	}

	return ctx.JSON(http.StatusOK, apiAccounts)
}

// GetAccount IDでアカウントを取得
func (s *Server) GetAccount(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Getting account by ID",
		logger.F("account_id", accountId),
	)

	account, err := s.accountUsecase.GetByID(reqCtx, accountId)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get account", err,
			logger.F("account_id", accountId),
		)
		return handleAccountError(ctx, err)
	}

	apiAccount := NewAPIAccountFromEntity(account)
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

	account, err := s.accountUsecase.Update(reqCtx, accountId, input)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to update account", err,
			logger.F("account_id", accountId),
		)
		return handleAccountError(ctx, err)
	}

	s.logger.Info(reqCtx, "Account updated successfully",
		logger.F("account_id", accountId),
	)

	apiAccount := NewAPIAccountFromEntity(account)
	return ctx.JSON(http.StatusOK, apiAccount)
}

// DeleteAccount アカウントを削除
func (s *Server) DeleteAccount(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Deleting account",
		logger.F("account_id", accountId),
	)

	err := s.accountUsecase.Delete(reqCtx, accountId)
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
