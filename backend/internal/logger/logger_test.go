package logger

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want slog.Level
	}{
		{name: "debug", in: "debug", want: slog.LevelDebug},
		{name: "info", in: "info", want: slog.LevelInfo},
		{name: "warn", in: "warn", want: slog.LevelWarn},
		{name: "warning も warn 扱い", in: "warning", want: slog.LevelWarn},
		{name: "error", in: "error", want: slog.LevelError},
		{name: "大文字小文字を無視", in: "DEBUG", want: slog.LevelDebug},
		{name: "前後の空白を無視", in: "  info  ", want: slog.LevelInfo},
		{name: "空文字は既定の info", in: "", want: slog.LevelInfo},
		{name: "未知の値は既定の info", in: "verbose", want: slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseLevel(tt.in); got != tt.want {
				t.Errorf("parseLevel(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
