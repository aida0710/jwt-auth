package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// ErrorHandler アプリケーション全体のエラーハンドラー
type ErrorHandler struct {
	stackSize int
}

// NewErrorHandler ErrorHandlerの新しいインスタンスを作成
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		stackSize: 4096,
	}
}

// HTTPErrorHandler EchoのHTTPエラーハンドラー
func (eh *ErrorHandler) HTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := "Internal Server Error"

	var he *echo.HTTPError
	if errors.As(err, &he) {
		code = he.Code
		message = fmt.Sprintf("%v", he.Message)
	}

	// ログレベルに応じた処理
	switch {
	case code == http.StatusNotFound || code == http.StatusMethodNotAllowed:
		c.Logger().Info("APIError: %v", err)
	case code >= 400 && code < 500:
		c.Logger().Warn("APIError: %v", err)
	case code >= 500:
		c.Logger().Error("APIError: %v", err)

		trace := make([]byte, eh.stackSize)
		n := runtime.Stack(trace, false)
		stackStr := string(trace[:n])

		c.Logger().Error("===== Start stack trace =====")
		c.Logger().Error(stackStr)
		c.Logger().Error("===== End stack trace =====")
	default:
		c.Logger().Error("APIError: %v", err)
	}

	// レスポンスがまだ送信されていない場合のみエラーレスポンスを送信
	if !c.Response().Committed {
		if err := c.JSON(code, map[string]interface{}{
			"error": message,
			"code":  code,
		}); err != nil {
			c.Logger().Error("Failed to send error response: %v", err)
		}
	}
}

// LoggingMiddleware エラーログ出力用のミドルウェア
func (eh *ErrorHandler) LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := next(c)
		if err != nil {
			// HTTPエラーの場合はHTTPErrorHandlerで処理するので、ここではスキップ
			var he *echo.HTTPError
			if errors.As(err, &he) {
				return err
			}

			trace := make([]byte, eh.stackSize)
			n := runtime.Stack(trace, false)
			stackStr := string(trace[:n])

			c.Logger().Error("RequestError: %v", err)
			c.Logger().Error("===== Start stack trace =====")
			c.Logger().Error(stackStr)
			c.Logger().Error("===== End stack trace =====")

			return err
		}
		return nil
	}
}

// RecoverConfig Recoverミドルウェアの設定を返す
func (eh *ErrorHandler) RecoverConfig() middleware.RecoverConfig {
	return middleware.RecoverConfig{
		DisablePrintStack: false,
		DisableStackAll:   false,
		LogLevel:          log.ERROR,
		StackSize:         eh.stackSize,
		LogErrorFunc:      eh.recoverLogError,
	}
}

// recoverLogError panicからのリカバリー時のログ出力
func (eh *ErrorHandler) recoverLogError(c echo.Context, err error, stack []byte) error {
	stackStr := string(stack)
	errMsg := fmt.Sprintf("RequestProcessingError: %v", err)

	c.Logger().Error(errMsg)
	c.Logger().Error("===== Start stack trace =====")
	c.Logger().Error(stackStr)
	c.Logger().Error("===== End stack trace =====")

	return c.JSON(http.StatusInternalServerError, map[string]string{
		"error": "Internal server error",
	})
}
