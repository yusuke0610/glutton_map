package pin

import (
	"context"
	"testing"
)

func TestSQLiteRepository_IPHashを保存しListForStatsで取り出せる(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	in := []Pin{
		{Prefecture: "東京都", Lat: 35.6, Lng: 139.7, IPHash: "hashA"}, // 同一ファン
		{Prefecture: "東京都", Lat: 35.7, Lng: 139.8, IPHash: "hashA"}, // の連投
		{Prefecture: "大阪府", Lat: 34.7, Lng: 135.5, IPHash: "hashB"},
	}
	for _, p := range in {
		if err := repo.Insert(ctx, p); err != nil {
			t.Fatalf("Insert(%v): %v", p, err)
		}
	}

	rows, err := repo.ListForStats(ctx)
	if err != nil {
		t.Fatalf("ListForStats: %v", err)
	}
	if len(rows) != len(in) {
		t.Fatalf("len(ListForStats) = %d, want %d（連投も保存されること）", len(rows), len(in))
	}

	// prefecture と ip_hash が往復して保存・取得できること。
	count := map[string]int{}
	for _, r := range rows {
		count[string(r.Prefecture)+"/"+r.IPHash]++
	}
	if count["東京都/hashA"] != 2 {
		t.Errorf("東京都/hashA = %d, want 2", count["東京都/hashA"])
	}
	if count["大阪府/hashB"] != 1 {
		t.Errorf("大阪府/hashB = %d, want 1", count["大阪府/hashB"])
	}
}
