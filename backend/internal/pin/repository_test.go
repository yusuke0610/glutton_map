package pin

import (
	"context"
	"path/filepath"
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
