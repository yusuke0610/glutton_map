package outbound

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler は計測付き送客リダイレクト /out を提供する素 Gin ハンドラ。
type Handler struct {
	repo ClickRepository
}

func NewHandler(repo ClickRepository) *Handler {
	return &Handler{repo: repo}
}

// Register は素 Gin ルートを登録する（strict-server とは別系統）。
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/out", h.out)
}

// out は to をホワイトリストで解決し、クリックを記録してから 302 で公式 URL へ送る。
// 未知の to は 404（任意 URL へはリダイレクトしない＝オープンリダイレクト防止）。
func (h *Handler) out(c *gin.Context) {
	to := c.Query("to")
	dest, ok := Resolve(to)
	if !ok {
		c.String(http.StatusNotFound, "リンクが見つかりません")
		return
	}

	click := Click{
		Destination: to,
		UTMSource:   c.Query("utm_source"),
		UTMMedium:   c.Query("utm_medium"),
		UTMCampaign: c.Query("utm_campaign"),
		CreatedAt:   time.Now().UTC(),
	}
	if err := h.repo.RecordClick(c.Request.Context(), click); err != nil {
		// 計測が落ちても送客は止めない（送客優先）。観測のためログだけ残す。
		slog.Error("送客クリックの記録に失敗", "to", to, "error", err)
	}

	c.Redirect(http.StatusFound, dest)
}
