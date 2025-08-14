package handler

import (
	"errors"
	"net/http"

	"github.com/aida0710/jwt-auth/internal/api"
	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/logger"
	"github.com/aida0710/jwt-auth/internal/usecase"
	"github.com/labstack/echo/v4"
)

// ====================================
// DTO変換関数
// ====================================

// NewAPIProjectFromEntity エンティティからAPIレスポンスに変換
func NewAPIProjectFromEntity(project *domain.Project) api.Project {
	apiProject := api.Project{
		Id:        project.ID,
		AccountId: project.AccountID,
		Name:      project.Name,
		CreatedAt: project.CreatedAt,
		UpdatedAt: project.UpdatedAt,
	}

	// Descriptionの変換
	if project.Description != "" {
		apiProject.Description = &project.Description
	}

	// Statusの変換
	switch project.Status {
	case domain.ProjectStatusActive:
		apiProject.Status = api.ProjectStatusActive
	case domain.ProjectStatusInactive:
		apiProject.Status = api.ProjectStatusInactive
	case domain.ProjectStatusArchived:
		apiProject.Status = api.ProjectStatusArchived
	default:
		apiProject.Status = api.ProjectStatusActive
	}

	return apiProject
}

// ====================================
// プロジェクト関連のハンドラー実装
// ====================================

// ListProjects アカウントのプロジェクト一覧を取得
func (s *Server) ListProjects(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Getting projects for account",
		logger.F("account_id", accountId),
	)

	projects, err := s.projectUsecase.ListByAccountID(reqCtx, accountId)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get projects", err,
			logger.F("account_id", accountId),
		)
		return handleProjectError(ctx, err)
	}

	// エンティティからAPIレスポンスに変換
	apiProjects := make([]api.Project, len(projects))
	for i, project := range projects {
		apiProjects[i] = NewAPIProjectFromEntity(project)
	}

	return ctx.JSON(http.StatusOK, apiProjects)
}

// CreateProject 新しいプロジェクトを作成
func (s *Server) CreateProject(ctx echo.Context, accountId api.AccountID) error {
	reqCtx := ctx.Request().Context()

	var req api.CreateProjectRequest
	if err := ctx.Bind(&req); err != nil {
		s.logger.Warn(reqCtx, "Invalid request body", logger.F("error", err.Error()))
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: "Invalid request body",
		})
	}

	s.logger.Info(reqCtx, "Creating new project",
		logger.F("account_id", accountId),
		logger.F("name", req.Name),
	)

	input := usecase.CreateProjectInput{
		Name: req.Name,
	}

	// Descriptionが提供されている場合
	if req.Description != nil {
		input.Description = *req.Description
	}

	// Statusが提供されている場合
	if req.Status != nil {
		status := string(*req.Status)
		input.Status = &status
	}

	project, err := s.projectUsecase.Create(reqCtx, accountId, input)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to create project", err,
			logger.F("account_id", accountId),
		)
		return handleProjectError(ctx, err)
	}

	s.logger.Info(reqCtx, "Project created successfully",
		logger.F("project_id", project.ID),
		logger.F("account_id", accountId),
	)

	apiProject := NewAPIProjectFromEntity(project)
	return ctx.JSON(http.StatusCreated, apiProject)
}

// GetProject IDでプロジェクトを取得
func (s *Server) GetProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Getting project by ID",
		logger.F("account_id", accountId),
		logger.F("project_id", projectId),
	)

	project, err := s.projectUsecase.GetByID(reqCtx, accountId, projectId)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to get project", err,
			logger.F("account_id", accountId),
			logger.F("project_id", projectId),
		)
		return handleProjectError(ctx, err)
	}

	apiProject := NewAPIProjectFromEntity(project)
	return ctx.JSON(http.StatusOK, apiProject)
}

// UpdateProject プロジェクトを更新
func (s *Server) UpdateProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error {
	reqCtx := ctx.Request().Context()

	var req api.UpdateProjectRequest
	if err := ctx.Bind(&req); err != nil {
		s.logger.Warn(reqCtx, "Invalid request body", logger.F("error", err.Error()))
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: "Invalid request body",
		})
	}

	s.logger.Info(reqCtx, "Updating project",
		logger.F("account_id", accountId),
		logger.F("project_id", projectId),
	)

	input := usecase.UpdateProjectInput{
		Name:        req.Name,
		Description: req.Description,
	}

	if req.Status != nil {
		status := string(*req.Status)
		input.Status = &status
	}

	project, err := s.projectUsecase.Update(reqCtx, accountId, projectId, input)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to update project", err,
			logger.F("account_id", accountId),
			logger.F("project_id", projectId),
		)
		return handleProjectError(ctx, err)
	}

	s.logger.Info(reqCtx, "Project updated successfully",
		logger.F("account_id", accountId),
		logger.F("project_id", projectId),
	)

	apiProject := NewAPIProjectFromEntity(project)
	return ctx.JSON(http.StatusOK, apiProject)
}

// DeleteProject プロジェクトを削除
func (s *Server) DeleteProject(ctx echo.Context, accountId api.AccountID, projectId api.ProjectID) error {
	reqCtx := ctx.Request().Context()

	s.logger.Info(reqCtx, "Deleting project",
		logger.F("account_id", accountId),
		logger.F("project_id", projectId),
	)

	err := s.projectUsecase.Delete(reqCtx, accountId, projectId)
	if err != nil {
		s.logger.Error(reqCtx, "Failed to delete project", err,
			logger.F("account_id", accountId),
			logger.F("project_id", projectId),
		)
		return handleProjectError(ctx, err)
	}

	s.logger.Info(reqCtx, "Project deleted successfully",
		logger.F("account_id", accountId),
		logger.F("project_id", projectId),
	)

	return ctx.NoContent(http.StatusNoContent)
}

// ====================================
// エラーハンドリング
// ====================================

// handleProjectError プロジェクト関連のエラーをHTTPレスポンスに変換
func handleProjectError(ctx echo.Context, err error) error {
	// エラーマッピングから適切なステータスコードを探す
	if errors.Is(err, domain.ErrProjectNotFound) || errors.Is(err, domain.ErrAccountNotFound) {
		return ctx.JSON(http.StatusNotFound, api.Error{
			Error: err.Error(),
		})
	}
	if errors.Is(err, domain.ErrProjectLimitExceeded) {
		return ctx.JSON(http.StatusConflict, api.Error{
			Error: err.Error(),
		})
	}
	if errors.Is(err, domain.ErrInvalidAccountID) || errors.Is(err, domain.ErrInvalidStatus) ||
		errors.Is(err, domain.ErrInvalidID) || errors.Is(err, domain.ErrInvalidName) {
		return ctx.JSON(http.StatusBadRequest, api.Error{
			Error: err.Error(),
		})
	}

	// デフォルトのエラーレスポンス
	return ctx.JSON(http.StatusInternalServerError, api.Error{
		Error: "Internal server error",
	})
}
