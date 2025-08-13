package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// GetHealth ヘルスチェックエンドポイント
func (s *Server) GetHealth(ctx echo.Context) error {
	s.logger.Debug(ctx.Request().Context(), "Health check requested")

	return ctx.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
