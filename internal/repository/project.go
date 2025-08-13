package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aida0710/jwt-auth/internal/domain"
	"github.com/aida0710/jwt-auth/internal/infrastructure/database"
	"github.com/jmoiron/sqlx"
)

// projectRepository repository.ProjectRepositoryの実装
type projectRepository struct {
	db *sqlx.DB
}

// NewProjectRepository 新しいプロジェクトリポジトリを作成
func NewProjectRepository(db *sqlx.DB) domain.ProjectRepository {
	return &projectRepository{
		db: db,
	}
}

// Create 新しいプロジェクトを作成
func (r *projectRepository) Create(ctx context.Context, project *domain.Project) error {
	query := `
		INSERT INTO projects (id, account_id, name, description, status, created_at, updated_at)
		VALUES (:id, :account_id, :name, :description, :status, :created_at, :updated_at)
	`

	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	_, err := exec.NamedExecContext(ctx, query, project)
	if err != nil {
		return err
	}

	return nil
}

// GetByID IDでプロジェクトを取得
func (r *projectRepository) GetByID(ctx context.Context, id string) (*domain.Project, error) {
	var project domain.Project
	query := `
		SELECT id, account_id, name, description, status, created_at, updated_at
		FROM projects
		WHERE id = ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &project, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &project, nil
}

// GetByAccountID アカウントIDでプロジェクトをページネーション付きで取得
func (r *projectRepository) GetByAccountID(ctx context.Context, accountID string, limit, offset int) ([]*domain.Project, error) {
	projects := make([]*domain.Project, 0)
	query := `
		SELECT id, account_id, name, description, status, created_at, updated_at
		FROM projects
		WHERE account_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.SelectContext(ctx, &projects, query, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// List すべてのプロジェクトをページネーション付きで取得
func (r *projectRepository) List(ctx context.Context, limit, offset int) ([]*domain.Project, error) {
	projects := make([]*domain.Project, 0)
	query := `
		SELECT id, account_id, name, description, status, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.SelectContext(ctx, &projects, query, limit, offset)
	if err != nil {
		return nil, err
	}

	return projects, nil
}

// Update プロジェクトを更新
func (r *projectRepository) Update(ctx context.Context, project *domain.Project) error {
	query := `
		UPDATE projects
		SET name = :name, description = :description, status = :status, updated_at = :updated_at
		WHERE id = :id
	`

	project.UpdatedAt = time.Now()

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.NamedExecContext(ctx, query, project)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrProjectNotFound
	}

	return nil
}

// Delete プロジェクトを削除
func (r *projectRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM projects WHERE id = ?`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	result, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrProjectNotFound
	}

	return nil
}

// DeleteByAccountID アカウントIDですべてのプロジェクトを削除
func (r *projectRepository) DeleteByAccountID(ctx context.Context, accountID string) error {
	query := `DELETE FROM projects WHERE account_id = ?`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	_, err := exec.ExecContext(ctx, query, accountID)
	if err != nil {
		return err
	}

	return nil
}

// CountByAccountID アカウントIDでプロジェクト数をカウント
func (r *projectRepository) CountByAccountID(ctx context.Context, accountID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE account_id = ?`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &count, query, accountID)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Count プロジェクトの総数をカウント
func (r *projectRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects`

	// トランザクション対応
	exec := database.GetExecutor(ctx, r.db)
	err := exec.GetContext(ctx, &count, query)
	if err != nil {
		return 0, err
	}

	return count, nil
}
