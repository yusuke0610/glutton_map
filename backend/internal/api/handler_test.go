package api

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// fakeRepo は PinRepository のテストダブル（フェイク）。
// DB を立てずに、返したいピンやエラーを差し込めるようにする。
type fakeRepo struct {
	pins []pin.Pin
	err  error
}

func (f *fakeRepo) GetPins(ctx context.Context) ([]pin.Pin, error) {
	return f.pins, f.err
}

// Count / Insert は handler のテストでは使わないので最小限のスタブ。
func (f *fakeRepo) Count(ctx context.Context) (int, error)   { return len(f.pins), nil }
func (f *fakeRepo) Insert(ctx context.Context, p pin.Pin) error { return nil }

func TestGetApiPins_集計して返す(t *testing.T) {
	// 東京2件・大阪1件 → 3都道府県ではなく2都道府県、総数3。
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "東京都", Lat: 35.6, Lng: 139.7},
		{Prefecture: "東京都", Lat: 35.7, Lng: 139.8},
		{Prefecture: "大阪府", Lat: 34.7, Lng: 135.5},
	}}
	h := NewHandler(repo)

	resp, err := h.GetApiPins(context.Background(), GetApiPinsRequestObject{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	got, ok := resp.(GetApiPins200JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型が想定外: %T", resp)
	}

	if got.PrefectureCount != 2 {
		t.Errorf("PrefectureCount = %d, want 2", got.PrefectureCount)
	}
	if got.Total != 3 {
		t.Errorf("Total = %d, want 3", got.Total)
	}
	if len(got.Pins) != 3 {
		t.Fatalf("len(Pins) = %d, want 3", len(got.Pins))
	}
	// 最小スライスでは weight は常に 1。
	for i, p := range got.Pins {
		if p.Weight == nil || *p.Weight != 1 {
			t.Errorf("Pins[%d].Weight = %v, want 1", i, p.Weight)
		}
	}
}

func TestGetApiPins_repoのエラーを伝播する(t *testing.T) {
	wantErr := errors.New("db 接続失敗")
	h := NewHandler(&fakeRepo{err: wantErr})

	_, err := h.GetApiPins(context.Background(), GetApiPinsRequestObject{})
	if err == nil {
		t.Fatal("err = nil, want エラー")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
	// 元エラーを %w で包みつつ、どの操作で失敗したかの文脈を付与する。
	if !strings.Contains(err.Error(), "ピン取得") {
		t.Errorf("err = %q, want に文脈 \"ピン取得\" を含む", err.Error())
	}
}
