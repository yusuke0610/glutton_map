package pin

import (
	"context"
	"testing"
	"time"

	"github.com/kisaragi-ai-map/backend/internal/outbound"
)

func TestSQLiteRepository_クリックを記録して往復する(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// 同じリポジトリが ClickRepository も満たす（DB 接続を1つに保つ）。
	clickRepo, ok := repo.(outbound.ClickRepository)
	if !ok {
		t.Fatalf("repo は outbound.ClickRepository を満たすべき: %T", repo)
	}

	at := time.Date(2026, 6, 24, 12, 0, 0, 0, time.UTC)
	want := outbound.Click{
		Destination: "official_menu",
		UTMSource:   "twitter", UTMMedium: "social", UTMCampaign: "fan_share",
		CreatedAt: at,
	}
	if err := clickRepo.RecordClick(ctx, want); err != nil {
		t.Fatalf("RecordClick: %v", err)
	}

	// 読み出しは同パッケージのテスト専用ヘルパで確認する（DB 隔離のためドライバはここに閉じる）。
	sr, ok := repo.(*sqliteRepo)
	if !ok {
		t.Fatalf("repo は *sqliteRepo のはず: %T", repo)
	}
	clicks, err := sr.listClicks(ctx)
	if err != nil {
		t.Fatalf("listClicks: %v", err)
	}
	if len(clicks) != 1 {
		t.Fatalf("len(clicks) = %d, want 1", len(clicks))
	}
	got := clicks[0]
	if got.Destination != want.Destination {
		t.Errorf("Destination = %q, want %q", got.Destination, want.Destination)
	}
	if got.UTMSource != "twitter" || got.UTMMedium != "social" || got.UTMCampaign != "fan_share" {
		t.Errorf("UTM = %q/%q/%q, want twitter/social/fan_share", got.UTMSource, got.UTMMedium, got.UTMCampaign)
	}
	// 時刻は UTC で往復すること。
	if !got.CreatedAt.UTC().Equal(at) {
		t.Errorf("CreatedAt = %v, want %v (UTC)", got.CreatedAt.UTC(), at)
	}
}
