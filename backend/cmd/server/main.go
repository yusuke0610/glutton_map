package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/kisaragi-ai-map/backend/internal/api"
	"github.com/kisaragi-ai-map/backend/internal/db"
	"github.com/kisaragi-ai-map/backend/internal/pin"
)

func main() {
	dsn := os.Getenv("LIBSQL_URL")
	if dsn == "" {
		dsn = "file:./data/pins.db"
	}

	repo, err := pin.NewSQLiteRepository(dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}

	// 起動時、DB が空なら seed を流す。
	if err := db.Seed(context.Background(), repo); err != nil {
		log.Fatalf("seed: %v", err)
	}

	// 許可するフロントのオリジン。環境変数 CORS_ALLOW_ORIGINS（カンマ区切り）で
	// 上書きでき、未設定ならローカル開発の既定値を使う。
	allowOrigins := []string{"http://localhost:5173"}
	if v := os.Getenv("CORS_ALLOW_ORIGINS"); v != "" {
		allowOrigins = strings.Split(v, ",")
	}

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowMethods: []string{"GET"},
		AllowHeaders: []string{"Origin", "Content-Type"},
		MaxAge:       12 * time.Hour,
	}))

	// strict-server: NewStrictHandler でラップしてから登録する。
	h := api.NewStrictHandler(api.NewHandler(repo), nil)
	api.RegisterHandlers(router, h)

	addr := ":8000"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	if err := router.Run(addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}
