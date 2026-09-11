package httpmw

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxRequestBodyBytes は POST /api/pins が許容する最大ボディサイズ。
// openapi.yaml の maxLength（nickname 30 / city 50 / comment 200 / utm等 各64）から
// 逆算しても数百バイトで足りるが、UTF-8 のマルチバイトと JSON のオーバーヘッドを
// 見込んで余裕を持たせる。
const MaxRequestBodyBytes = 16 * 1024

// MaxBodyBytes は c.Request.Body を http.MaxBytesReader でラップし、
// 上限を超えるボディが JSON パース前に拒否されるようにする。
// 認証なしの公開 POST エンドポイントでの無制限メモリ確保（DoS）を防ぐ。
func MaxBodyBytes(max int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, max)
		c.Next()
	}
}
