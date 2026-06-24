package outbound

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

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

func TestOut_長すぎるUTMは切り詰めて記録する(t *testing.T) {
	repo := &fakeClickRepo{}
	r := newRouter(repo)

	long := strings.Repeat("a", 100) // maxUTMLen(64) 超え
	req := httptest.NewRequest(http.MethodGet, "/out?to=official_menu&utm_source="+long, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302", w.Code)
	}
	if len(repo.recorded) != 1 {
		t.Fatalf("recorded = %d件, want 1", len(repo.recorded))
	}
	if n := utf8.RuneCountInString(repo.recorded[0].UTMSource); n != maxUTMLen {
		t.Errorf("UTMSource 長 = %d, want %d（切り詰め）", n, maxUTMLen)
	}
}

func TestOut_同一IP連打は二重計測しないが送客は通す(t *testing.T) {
	repo := &fakeClickRepo{}
	r := newRouter(repo)

	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/out?to=official_menu", nil)
		req.RemoteAddr = "203.0.113.9:5555" // 同一 IP からの連打
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	// 1回目・2回目とも送客（302）は通るが、計測は1件だけ（連打の水増しを抑制）。
	if c := send(); c != http.StatusFound {
		t.Fatalf("1回目 status = %d, want 302", c)
	}
	if c := send(); c != http.StatusFound {
		t.Fatalf("2回目 status = %d, want 302（送客は止めない）", c)
	}
	if len(repo.recorded) != 1 {
		t.Errorf("recorded = %d件, want 1（同一IP連打は二重計測しない）", len(repo.recorded))
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
