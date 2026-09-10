package api

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// jsonError は Error スキーマ（{"message": ...}）に準拠したエラーレスポンスを返す。
// Go の内部エラー文字列は外部に出さず、原因は slog へ記録する。
func jsonError(c *gin.Context, status int, publicMessage string, err error) {
	if err != nil {
		slog.Error(publicMessage, "error", err, "status", status)
	}
	c.JSON(status, gin.H{"message": publicMessage})
}

// DefaultGinServerOptions は生成コードの既定 ErrorHandler（{"msg": ...}）を
// 契約準拠のレスポンスへ差し替える GinServerOptions を返す。
// クエリパラメータのバインド失敗（例: lat=abc）などで使われる。
func DefaultGinServerOptions() GinServerOptions {
	return GinServerOptions{
		ErrorHandler: func(c *gin.Context, err error, statusCode int) {
			jsonError(c, statusCode, "リクエストのパラメータが不正です", err)
		},
	}
}

// DefaultStrictGinServerOptions は strict-server の既定エラーハンドラ
// （{"msg": ...}）を契約準拠のレスポンスへ差し替える StrictGinServerOptions を返す。
func DefaultStrictGinServerOptions() StrictGinServerOptions {
	return StrictGinServerOptions{
		RequestErrorHandlerFunc: func(c *gin.Context, err error) {
			jsonError(c, http.StatusBadRequest, "リクエストの解析に失敗しました", err)
		},
		HandlerErrorFunc: func(c *gin.Context, err error) {
			jsonError(c, http.StatusInternalServerError, "サーバエラーが発生しました", err)
		},
		ResponseErrorHandlerFunc: func(c *gin.Context, err error) {
			jsonError(c, http.StatusInternalServerError, "サーバエラーが発生しました", err)
		},
	}
}
