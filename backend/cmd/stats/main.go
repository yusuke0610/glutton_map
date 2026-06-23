// Command stats は提出用のユニークファン集計レポートを標準出力に書き出す。
// 地図API（cmd/server）とは別経路で DB を読み、同一 ip_hash の連投・curl を
// 重複排除した「ファン数」を JSON（既定）または CSV で出力する。
//
// 使い方（Makefile 経由）: make stats             → JSON サマリ
//
//	make stats FORMAT=csv  → 都道府県別 CSV
//
// 出力をファイルに保存して提出する: make stats > report.json
package main

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/kisaragi-ai-map/backend/internal/pin"
	"github.com/kisaragi-ai-map/backend/internal/stats"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "stats:", err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("LIBSQL_URL")
	if dsn == "" {
		dsn = "file:./data/pins.db"
	}

	repo, err := pin.NewSQLiteRepository(dsn)
	if err != nil {
		return fmt.Errorf("DB を開けません: %w", err)
	}

	rows, err := repo.ListForStats(context.Background())
	if err != nil {
		return fmt.Errorf("集計データの取得: %w", err)
	}
	report := stats.Build(rows)

	if os.Getenv("FORMAT") == "csv" {
		return writeCSV(os.Stdout, report)
	}
	return writeJSON(os.Stdout, report)
}

func writeJSON(w *os.File, r stats.Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("JSON 出力: %w", err)
	}
	return nil
}

// writeCSV は都道府県別のユニークファン数を CSV で出力する（Excel で開きやすい）。
func writeCSV(w *os.File, r stats.Report) error {
	cw := csv.NewWriter(w)
	if err := cw.Write([]string{"prefecture", "unique_fans"}); err != nil {
		return fmt.Errorf("CSV 出力: %w", err)
	}
	prefs := make([]string, 0, len(r.ByPrefecture))
	for p := range r.ByPrefecture {
		prefs = append(prefs, p)
	}
	sort.Strings(prefs) // 出力順を安定させる
	for _, p := range prefs {
		if err := cw.Write([]string{p, strconv.Itoa(r.ByPrefecture[p])}); err != nil {
			return fmt.Errorf("CSV 出力: %w", err)
		}
	}
	cw.Flush()
	return cw.Error()
}
