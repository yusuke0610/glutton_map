package outbound

import (
	"log/slog"
	"net/http"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/kisaragi-ai-map/backend/internal/httpmw"
)

// maxUTMLen は記録する UTM 値の上限。無認証 GET なので任意長の値で outbound_clicks を
// 肥大化させられないよう、保存前に切り詰める（openapi の utm maxLength と揃える）。
const maxUTMLen = 64

// recordCooldown は同一 IP の連打を計測対象から外すクールダウン。送客自体は止めない。
const recordCooldown = 3 * time.Second

// Handler は計測付き送客リダイレクト /out を提供する素 Gin ハンドラ。
type Handler struct {
	repo ClickRepository
	// limiter は同一 IP の短時間連打を「計測しない」ためのゲート（送客は常に通す）。
	// 無認証 GET によるメトリクス水増し・行の無制限増加を抑える。
	limiter *httpmw.Limiter
}

func NewHandler(repo ClickRepository) *Handler {
	return &Handler{repo: repo, limiter: httpmw.NewLimiter(recordCooldown)}
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

	// 同一 IP の短時間連打は計測しない（メトリクス水増し・行の無制限増加の抑制）。
	// 送客自体は止めないため、クールダウン中でもこの後の Redirect は必ず行う。
	if h.limiter.Allow(c.ClientIP()) {
		click := Click{
			Destination: to,
			// UTM は無認証クエリ由来なので保存前に上限まで切り詰める。
			UTMSource:   capLen(c.Query("utm_source"), maxUTMLen),
			UTMMedium:   capLen(c.Query("utm_medium"), maxUTMLen),
			UTMCampaign: capLen(c.Query("utm_campaign"), maxUTMLen),
			CreatedAt:   time.Now().UTC(),
		}
		if err := h.repo.RecordClick(c.Request.Context(), click); err != nil {
			// 計測が落ちても送客は止めない（送客優先）。観測のためログだけ残す。
			slog.Error("送客クリックの記録に失敗", "to", to, "error", err)
		}
	}

	c.Redirect(http.StatusFound, dest)
}

// capLen は文字列を最大 n ルーンまで切り詰める（マルチバイト境界を壊さない）。
func capLen(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}
