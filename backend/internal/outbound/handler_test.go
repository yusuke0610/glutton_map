package outbound

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeClickRepo は ClickRepository のテストダブル。
type fakeClickRepo struct {
	recorded []Click
	err      error
}

func (f *fakeClickRepo) RecordClick(_ context.Context, c Click) error {
	if f.err != nil {
		return f.err
	}
	f.recorded = append(f.recorded, c)
	return nil
}

func newRouter(repo ClickRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(repo).Register(r)
	return r
}

func TestOut_既知のキーは記録して302(t *testing.T) {
	repo := &fakeClickRepo{}
	r := newRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/out?to=official_menu&utm_source=map&utm_medium=cta", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 (body=%s)", w.Code, w.Body.String())
	}
	if loc := w.Header().Get("Location"); loc != Destinations["official_menu"] {
		t.Errorf("Location = %q, want %q", loc, Destinations["official_menu"])
	}
	if len(repo.recorded) != 1 {
		t.Fatalf("recorded = %d件, want 1", len(repo.recorded))
	}
	got := repo.recorded[0]
	if got.Destination != "official_menu" {
		t.Errorf("Destination = %q, want official_menu", got.Destination)
	}
	if got.UTMSource != "map" || got.UTMMedium != "cta" {
		t.Errorf("UTM = %q/%q, want map/cta", got.UTMSource, got.UTMMedium)
	}
	// 時刻は UTC で記録すること。
	if got.CreatedAt.IsZero() || got.CreatedAt.Location() != nil && got.CreatedAt.Location().String() != "UTC" {
		t.Errorf("CreatedAt = %v, want 非ゼロの UTC", got.CreatedAt)
	}
}

func TestOut_未知のキーは404でリダイレクトしない(t *testing.T) {
	repo := &fakeClickRepo{}
	r := newRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/out?to=evil_target", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, want 空（オープンリダイレクト防止）", loc)
	}
	if len(repo.recorded) != 0 {
		t.Errorf("recorded = %d件, want 0", len(repo.recorded))
	}
}

func TestOut_to未指定は404(t *testing.T) {
	r := newRouter(&fakeClickRepo{})
	req := httptest.NewRequest(http.MethodGet, "/out", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestOut_計測失敗でも送客は止めない(t *testing.T) {
	// 計測(DB)が落ちても 302 は通す（送客優先）。
	repo := &fakeClickRepo{err: errors.New("db down")}
	r := newRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/out?to=official_menu", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302（計測失敗でも送客）", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != Destinations["official_menu"] {
		t.Errorf("Location = %q, want %q", loc, Destinations["official_menu"])
	}
}
