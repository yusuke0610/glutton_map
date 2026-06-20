package db

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// memRepo は挿入を記録するインメモリのフェイク。Seed のロジック検証用。
type memRepo struct {
	pins []pin.Pin
}

func (m *memRepo) GetPins(ctx context.Context) ([]pin.Pin, error) { return m.pins, nil }
func (m *memRepo) Count(ctx context.Context) (int, error)         { return len(m.pins), nil }
func (m *memRepo) Insert(ctx context.Context, p pin.Pin) error {
	m.pins = append(m.pins, p)
	return nil
}

// failRepo は指定した操作でエラーを返すフェイク。エラーの文脈付与を検証する。
type failRepo struct {
	failCount  bool
	failInsert bool
	err        error
}

func (f *failRepo) GetPins(ctx context.Context) ([]pin.Pin, error) { return nil, nil }
func (f *failRepo) Count(ctx context.Context) (int, error) {
	if f.failCount {
		return 0, f.err
	}
	return 0, nil
}
func (f *failRepo) Insert(ctx context.Context, p pin.Pin) error {
	if f.failInsert {
		return f.err
	}
	return nil
}

// countByPrefecture は都道府県ごとの件数を数える。
func countByPrefecture(pins []pin.Pin) map[pin.Prefecture]int {
	out := map[pin.Prefecture]int{}
	for _, p := range pins {
		out[p.Prefecture]++
	}
	return out
}

func TestSeed_空のDBに投入する(t *testing.T) {
	repo := &memRepo{}
	if err := Seed(context.Background(), repo); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	if len(repo.pins) == 0 {
		t.Fatal("Seed 後もピンが0件")
	}

	counts := countByPrefecture(repo.pins)
	// seedWeights に従う件数。
	if got := counts["高知県"]; got != 14 {
		t.Errorf("高知県 = %d, want 14", got)
	}
	if got := counts["東京都"]; got != 10 {
		t.Errorf("東京都 = %d, want 10", got)
	}
	// seedWeights 未指定の県は既定の 1 件。
	if got := counts["秋田県"]; got != 1 {
		t.Errorf("秋田県 = %d, want 1", got)
	}
}

func TestSeed_既にデータがあれば何もしない(t *testing.T) {
	repo := &memRepo{pins: []pin.Pin{{Prefecture: "東京都", Lat: 35.69, Lng: 139.69}}}

	if err := Seed(context.Background(), repo); err != nil {
		t.Fatalf("Seed: %v", err)
	}

	if len(repo.pins) != 1 {
		t.Errorf("ピン数 = %d, want 1（投入されないはず）", len(repo.pins))
	}
}

func TestSeed_Countエラーに文脈を付与する(t *testing.T) {
	wantErr := errors.New("db count 失敗")
	repo := &failRepo{failCount: true, err: wantErr}

	err := Seed(context.Background(), repo)
	if err == nil {
		t.Fatal("err = nil, want エラー")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v を包む", err, wantErr)
	}
	if !strings.Contains(err.Error(), "seed") {
		t.Errorf("err = %q, want に文脈 \"seed\" を含む", err.Error())
	}
}

func TestSeed_Insertエラーに文脈を付与する(t *testing.T) {
	wantErr := errors.New("db insert 失敗")
	repo := &failRepo{failInsert: true, err: wantErr}

	err := Seed(context.Background(), repo)
	if err == nil {
		t.Fatal("err = nil, want エラー")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v を包む", err, wantErr)
	}
	if !strings.Contains(err.Error(), "seed") {
		t.Errorf("err = %q, want に文脈 \"seed\" を含む", err.Error())
	}
}

func TestSeed_冪等(t *testing.T) {
	repo := &memRepo{}
	ctx := context.Background()

	if err := Seed(ctx, repo); err != nil {
		t.Fatalf("Seed 1回目: %v", err)
	}
	after1 := len(repo.pins)

	if err := Seed(ctx, repo); err != nil {
		t.Fatalf("Seed 2回目: %v", err)
	}
	after2 := len(repo.pins)

	if after1 != after2 {
		t.Errorf("2回目で件数が変化: %d → %d（冪等であるべき）", after1, after2)
	}
}
