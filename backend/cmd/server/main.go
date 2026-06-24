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
	"github.com/kisaragi-ai-map/backend/internal/share"
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
	// ミドルウェアが request context に載せた値（ip_hash）を、strict-server の
	// ハンドラが context.Context 経由で読めるようにする。
	router.ContextWithFallback = true

	// ClientIP() の信頼境界を明示する。未設定だと gin は全プロキシを信頼し
	// X-Forwarded-For を無検証で採用するため、ヘッダ偽装でレート制限や ip_hash の
	// 重複排除を回避できてしまう。既定は localhost のみ信頼（直公開でも詐称不可）。
	// 実際にリバースプロキシ/LB の背後に置く場合は TRUSTED_PROXIES でその IP を指定する。
	trustedProxies := []string{"127.0.0.1", "::1"}
	if v := os.Getenv("TRUSTED_PROXIES"); v != "" {
		trustedProxies = strings.Split(v, ",")
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		log.Error("invalid TRUSTED_PROXIES", "error", err)
		os.Exit(1)
	}

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

	// 投稿者の匿名識別子（ip_hash）を context に載せる。投稿は拒否せず、提出用集計で
	// 連投・curl をユニーク化するために使う。salt は固定値を使うこと（変えると過去ハッシュと
	// 一致しなくなる）。生IPは保存しない。
	ipSalt := os.Getenv("IP_HASH_SALT")
	if ipSalt == "" {
		ipSalt = "glutton-map-dev-salt" // 開発用デフォルト。本番は IP_HASH_SALT を必ず設定する。
		log.Warn("IP_HASH_SALT 未設定: 開発用デフォルトを使用（本番では必ず設定すること）")
	}
	router.Use(httpmw.IPHashMiddleware(ipSalt))

	// strict-server: NewStrictHandler でラップしてから登録する。
	h := api.NewStrictHandler(api.NewHandler(repo), nil)
	api.RegisterHandlers(router, h)

	// X 共有用の SSR ルート（/share, /static/ogp.png）。JSON ではないため strict-server には
	// 乗せず、素の Gin ルートとして登録する。og:image/og:url の絶対化に PUBLIC_BASE_URL、
	// 人間向けの着地先に FRONTEND_BASE_URL を使う（未設定はローカル開発の既定値）。
	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")
	if publicBaseURL == "" {
		publicBaseURL = "http://localhost:8001"
	}
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "http://localhost:5174"
	}
	share.NewHandler(share.Config{
		PublicBaseURL:   publicBaseURL,
		FrontendBaseURL: frontendBaseURL,
	}).Register(router)

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
