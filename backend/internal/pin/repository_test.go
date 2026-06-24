package pin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestRepo は t.TempDir() 上の使い捨て SQLite に接続した本物のリポジトリを返す。
// フェイクではなく実 DB を通すことで、スキーマ・マッピング・AutoMigrate まで検証する。
func newTestRepo(t *testing.T) PinRepository {
	t.Helper()
	dsn := "file:" + filepath.Join(t.TempDir(), "test.db")
	repo, err := NewSQLiteRepository(dsn)
	if err != nil {
		t.Fatalf("NewSQLiteRepository: %v", err)
	}
	return repo
}

func TestNewSQLiteRepository_不正DSNは文脈付きエラー(t *testing.T) {
	// 通常ファイルを作り、その配下を DB パスに指定して open/AutoMigrate を失敗させる。
	f := filepath.Join(t.TempDir(), "afile")
	if err := os.WriteFile(f, []byte("x"), 0o600); err != nil {
		t.Fatalf("前提のファイル作成に失敗: %v", err)
	}
	dsn := "file:" + filepath.Join(f, "x.db") // ファイル配下なのでディレクトリとして開けない

	_, err := NewSQLiteRepository(dsn)
	if err == nil {
		t.Fatal("不正 DSN なのにエラーが返らない")
	}
	// 接続・マイグレーションどちらの段で落ちても、自前の文脈が付いていること。
	if !strings.Contains(err.Error(), "DB 接続") && !strings.Contains(err.Error(), "マイグレーション") {
		t.Errorf("err = %q, want に文脈（\"DB 接続\" か \"マイグレーション\"）を含む", err.Error())
	}
}

func TestSQLiteRepository_新規DBは空(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	n, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 0 {
		t.Errorf("Count = %d, want 0", n)
	}

	pins, err := repo.GetPins(ctx)
	if err != nil {
		t.Fatalf("GetPins: %v", err)
	}
	if len(pins) != 0 {
		t.Errorf("len(GetPins) = %d, want 0", len(pins))
	}
}

func TestSQLiteRepository_InsertしたピンをGetPinsで取り出せる(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	want := Pin{Prefecture: "東京都", Lat: 35.69, Lng: 139.69}
	if err := repo.Insert(ctx, want); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	n, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Count: %v", err)
	}
	if n != 1 {
		t.Fatalf("Count = %d, want 1", n)
	}

	pins, err := repo.GetPins(ctx)
	if err != nil {
		t.Fatalf("GetPins: %v", err)
	}
	if len(pins) != 1 {
		t.Fatalf("len(GetPins) = %d, want 1", len(pins))
	}
	// 値が往復で欠けたり化けたりしないこと（pinRow ↔ Pin マッピングの検証）。
	got := pins[0]
	if got.Prefecture != want.Prefecture || got.Lat != want.Lat || got.Lng != want.Lng {
		t.Errorf("GetPins[0] = %+v, want %+v", got, want)
	}
}

func TestSQLiteRepository_投稿フィールドも往復する(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	want := Pin{
		Prefecture: "高知県", Lat: 33.56, Lng: 133.53,
		Nickname: "如月ファン", City: "高知市", Comment: "ここの唐揚げ弁当が最高",
	}
	if err := repo.Insert(ctx, want); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	pins, err := repo.GetPins(ctx)
	if err != nil {
		t.Fatalf("GetPins: %v", err)
	}
	if len(pins) != 1 {
		t.Fatalf("len(GetPins) = %d, want 1", len(pins))
	}
	got := pins[0]
	if got.Nickname != want.Nickname || got.City != want.City || got.Comment != want.Comment {
		t.Errorf("GetPins[0] = %+v, want nickname/city/comment = %q/%q/%q",
			got, want.Nickname, want.City, want.Comment)
	}
}

func TestSQLiteRepository_分析用フィールドも往復する(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	// 後から復元不可能な流入元・投稿時刻系。投稿の瞬間に保存され、往復で欠けないこと。
	want := Pin{
		Prefecture: "高知県", Lat: 33.56, Lng: 133.53,
		PrefectureCode: "39", AnonToken: "anon-xyz",
		UTMSource: "twitter", UTMMedium: "social", UTMCampaign: "fan_share",
	}
	if err := repo.Insert(ctx, want); err != nil {
		t.Fatalf("Insert: %v", err)
	}

	pins, err := repo.GetPins(ctx)
	if err != nil {
		t.Fatalf("GetPins: %v", err)
	}
	if len(pins) != 1 {
		t.Fatalf("len(GetPins) = %d, want 1", len(pins))
	}
	got := pins[0]
	if got.PrefectureCode != want.PrefectureCode {
		t.Errorf("PrefectureCode = %q, want %q", got.PrefectureCode, want.PrefectureCode)
	}
	if got.AnonToken != want.AnonToken {
		t.Errorf("AnonToken = %q, want %q", got.AnonToken, want.AnonToken)
	}
	if got.UTMSource != want.UTMSource || got.UTMMedium != want.UTMMedium || got.UTMCampaign != want.UTMCampaign {
		t.Errorf("UTM = %q/%q/%q, want %q/%q/%q",
			got.UTMSource, got.UTMMedium, got.UTMCampaign,
			want.UTMSource, want.UTMMedium, want.UTMCampaign)
	}
}

func TestSQLiteRepository_複数Insertを全件返す(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	in := []Pin{
		{Prefecture: "東京都", Lat: 35.69, Lng: 139.69},
		{Prefecture: "大阪府", Lat: 34.69, Lng: 135.50},
		{Prefecture: "高知県", Lat: 33.56, Lng: 133.53},
	}
	for _, p := range in {
		if err := repo.Insert(ctx, p); err != nil {
			t.Fatalf("Insert(%v): %v", p, err)
		}
	}

	pins, err := repo.GetPins(ctx)
	if err != nil {
		t.Fatalf("GetPins: %v", err)
	}
	if len(pins) != len(in) {
		t.Errorf("len(GetPins) = %d, want %d", len(pins), len(in))
	}
}
