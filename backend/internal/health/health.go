// Package health は運用系のヘルスチェックエンドポイントを提供する。
// JSON ではあるが openapi.yaml の公開契約（フロントが参照する API）ではないため、
// /share・/out と同様に strict-server には乗せず素の Gin ルートとして登録する。
package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Pinger は DB 疎通確認だけを行う最小の口。pin.PinRepository はこれを満たす。
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler は /healthz を提供する。
type Handler struct {
	pinger           Pinger
	readinessTimeout time.Duration
}

func NewHandler(pinger Pinger, readinessTimeout time.Duration) *Handler {
	return &Handler{pinger: pinger, readinessTimeout: readinessTimeout}
}

// Register は GET /healthz を登録する。
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/healthz", h.healthz)
}

// healthz は DB へ疎通確認(SELECT 1 相当)を行う readiness チェック。
// 疎通できなければ 503 を返す（起動直後や DB 障害時にロードバランサから外れるように
// するため、意図的に 200 ではなく 503 を選ぶ）。
func (h *Handler) healthz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.readinessTimeout)
	defer cancel()

	if err := h.pinger.Ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Check は url へ GET し、200 以外またはエラーなら失敗として返す。
// FROM scratch のイメージにはシェルも curl も無いため、コンテナの healthcheck は
// このサーバ自身のバイナリ（`server healthcheck` サブコマンド）から呼ぶ想定。
func Check(ctx context.Context, client *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("healthcheck リクエストの作成: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("healthcheck リクエスト: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: status = %d, want 200", resp.StatusCode)
	}
	return nil
}
