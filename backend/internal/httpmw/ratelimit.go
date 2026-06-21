// Package httpmw は HTTP の横断的関心事（ミドルウェア）を置く。
package httpmw

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Limiter は IP 単位の最小投稿間隔（クールダウン）を強制する素朴なメモリ内リミッタ。
// ファンマップの軽いスパム対策が目的で、分散環境やバースト許容は想定しない（最小スコープ）。
type Limiter struct {
	mu       sync.Mutex
	last     map[string]time.Time
	interval time.Duration
	now      func() time.Time // テストで差し替え可能な時計
}

// NewLimiter は指定間隔のリミッタを作る。
func NewLimiter(interval time.Duration) *Limiter {
	return &Limiter{
		last:     map[string]time.Time{},
		interval: interval,
		now:      time.Now,
	}
}

// Allow は ip からのアクセスを許可してよければ true を返し、許可時は最終時刻を更新する。
// 前回から interval 未満なら false（クールダウン中）。
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	// クールダウンを過ぎたエントリを掃除し、ローテーションする IP でマップが
	// 無制限に肥大化するのを防ぐ。件数が少ない前提の素朴な全走査で十分。
	for k, ts := range l.last {
		if now.Sub(ts) >= l.interval {
			delete(l.last, k)
		}
	}
	if last, ok := l.last[ip]; ok && now.Sub(last) < l.interval {
		return false
	}
	l.last[ip] = now
	return true
}

// Middleware は POST など書き込み系リクエストにクールダウンを課す gin ミドルウェア。
// 制限対象のメソッドのみ Limiter を適用し、超過時は 429 を返す。
func (l *Limiter) Middleware(methods ...string) gin.HandlerFunc {
	target := map[string]bool{}
	for _, m := range methods {
		target[m] = true
	}
	return func(c *gin.Context) {
		if target[c.Request.Method] && !l.Allow(c.ClientIP()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "投稿の間隔が短すぎます。少し待ってから再度お試しください。",
			})
			return
		}
		c.Next()
	}
}
