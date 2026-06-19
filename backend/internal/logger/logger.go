// Package logger はアプリ全体で使う構造化ロガー（log/slog ベース）を提供する。
package logger

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

// New は環境変数からレベルを決めた JSON ロガーを w 宛に作る。
// LOG_LEVEL: debug / info / warn / error（未設定・未知なら info）。
func New(w io.Writer) *slog.Logger {
	level := parseLevel(os.Getenv("LOG_LEVEL"))
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}

// parseLevel は文字列を slog.Level に変換する。大文字小文字・前後空白は無視し、
// 空文字や未知の値は既定の info にフォールバックする。
func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
