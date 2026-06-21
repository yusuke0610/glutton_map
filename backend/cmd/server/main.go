package main

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kisaragi-ai-map/backend/internal/api"
	"github.com/kisaragi-ai-map/backend/internal/db"
	"github.com/kisaragi-ai-map/backend/internal/httpmw"
	"github.com/kisaragi-ai-map/backend/internal/logger"
	"github.com/kisaragi-ai-map/backend/internal/pin"
)

func main() {
	// 構造化ロガーを用意し、標準 slog のデフォルトにも設定する。
	log := logger.New(os.Stdout)
	slog.SetDefault(log)

	dsn := os.Getenv("LIBSQL_URL")
	if dsn == "" {
		dsn = "file:./data/pins.db"
	}

	repo, err := pin.NewSQLiteRepository(dsn)
	if err != nil {
		log.Error("open db", "error", err)
		os.Exit(1)
	}

	// 初期ダミーデータの seed は既定で無効（実データ運用に移行したため）。
	// SEED_ON_START=true のときだけ DB が空なら投入する（E2E など用途限定）。
	if os.Getenv("SEED_ON_START") == "true" {
		if err := db.Seed(context.Background(), repo); err != nil {
			log.Error("seed", "error", err)
			os.Exit(1)
		}
	}

	// 許可するフロントのオリジン。環境変数 CORS_ALLOW_ORIGINS（カンマ区切り）で
	// 上書きでき、未設定ならローカル開発の既定値を使う。
	allowOrigins := []string{"http://localhost:5174"}
	if v := os.Getenv("CORS_ALLOW_ORIGINS"); v != "" {
		allowOrigins = strings.Split(v, ",")
	}

	// gin.Default() は標準の text ロガーを含むため、gin.New() に
	// Recovery と自前の slog リクエストログを載せて構造化ログに統一する。
	router := gin.New()
	router.Use(gin.Recovery(), requestLogger(log))
	router.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Origin", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	// 投稿(POST)のスパム対策: IP 単位のクールダウン。認証なしの軽い濫用対策。
	limiter := httpmw.NewLimiter(3 * time.Second)
	router.Use(limiter.Middleware("POST"))

	// strict-server: NewStrictHandler でラップしてから登録する。
	h := api.NewStrictHandler(api.NewHandler(repo), nil)
	api.RegisterHandlers(router, h)

	addr := ":8001"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Info("server starting", "addr", addr)
	if err := router.Run(addr); err != nil {
		log.Error("run", "error", err)
		os.Exit(1)
	}
}

// requestLogger は1リクエストごとに method/path/status/latency を構造化ログに出す。
func requestLogger(l *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		l.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
