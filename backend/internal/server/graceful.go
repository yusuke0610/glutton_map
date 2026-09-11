// Package server は http.Server のグレースフルシャットダウンを提供する。
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// Run は ln で srv を待ち受け、ctx がキャンセルされたら進行中のリクエストを
// shutdownTimeout の猶予内で完了させてから終了する。猶予を過ぎても終わらない
// リクエストがあれば、無限に待たず強制的に接続を閉じてエラーを返す。
func Run(ctx context.Context, srv *http.Server, ln net.Listener, shutdownTimeout time.Duration, log *slog.Logger) error {
	serveErrCh := make(chan error, 1)
	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
		serveErrCh <- nil
	}()

	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
		log.Info("shutting down", "shutdown_timeout", shutdownTimeout.String())
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			// 猶予を過ぎても終わらない接続を強制的に閉じ、無限待ちを避ける。
			_ = srv.Close()
			return fmt.Errorf("graceful shutdown: %w", err)
		}
		return nil
	}
}
