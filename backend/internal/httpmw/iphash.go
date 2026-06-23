package httpmw

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// ipHashContextKey は context に ip_hash を載せるための非公開キー型。
// 文字列キーの衝突を避けるため専用の型を使う。
type ipHashContextKey struct{}

// HashIP は生IPを salt 付き SHA-256 で一方向ハッシュ化する。
// 生IPは保存せず、分析用の匿名識別子（ip_hash）としてのみ扱うための変換。
// salt が変わると過去のハッシュと一致しなくなるため、運用では固定値を使うこと。
func HashIP(salt, ip string) string {
	sum := sha256.Sum256([]byte(salt + "\x00" + ip))
	return hex.EncodeToString(sum[:])
}

// withIPHash は ip_hash を context に載せて返す。
func withIPHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, ipHashContextKey{}, hash)
}

// IPHashFrom は context から ip_hash を取り出す。無ければ空文字。
// strict-server のハンドラは gin.Context を context.Context として受け取り、
// gin.Context.Value は未知キーを request の context へ委譲するため、ここで読める。
func IPHashFrom(ctx context.Context) string {
	if v, ok := ctx.Value(ipHashContextKey{}).(string); ok {
		return v
	}
	return ""
}

// IPHashMiddleware は salt 付きでクライアントIPをハッシュ化し、request の context に載せる。
// 生IPはどこにも保存せず、後続のハンドラ/リポジトリへは ip_hash だけが渡る。
func IPHashMiddleware(salt string) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := HashIP(salt, c.ClientIP())
		c.Request = c.Request.WithContext(withIPHash(c.Request.Context(), hash))
		c.Next()
	}
}
