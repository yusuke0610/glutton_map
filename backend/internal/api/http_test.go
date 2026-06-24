package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kisaragi-ai-map/backend/internal/httpmw"
	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// newTestRouter は main と同じ要領で strict-server を登録した gin エンジンを返す。
func newTestRouter(repo pin.PinRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewStrictHandler(NewHandler(repo), nil)
	RegisterHandlers(r, h)
	return r
}

func TestHTTP_GetApiPins_200とJSON契約(t *testing.T) {
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "東京都", Lat: 35.69, Lng: 139.69},
		{Prefecture: "東京都", Lat: 35.70, Lng: 139.80},
		{Prefecture: "大阪府", Lat: 34.69, Lng: 135.50},
	}}
	r := newTestRouter(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/pins", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	// 生成型 PinsResponse にデコードできること＝openapi.yaml の契約に沿っていること。
	var body PinsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON decode: %v (body=%s)", err, w.Body.String())
	}

	if body.PrefectureCount != 2 {
		t.Errorf("prefecture_count = %d, want 2", body.PrefectureCount)
	}
	if body.Total != 3 {
		t.Errorf("total = %d, want 3", body.Total)
	}
	if len(body.Pins) != 3 {
		t.Errorf("len(pins) = %d, want 3", len(body.Pins))
	}
}

func TestHTTP_GetPrefectureAt_200とJSON契約(t *testing.T) {
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "東京都", Lat: 35.69, Lng: 139.69},
		{Prefecture: "東京都", Lat: 35.70, Lng: 139.80},
		{Prefecture: "大阪府", Lat: 34.69, Lng: 135.50},
	}}
	r := newTestRouter(repo)

	// 練馬区あたり（東京都）の座標。
	req := httptest.NewRequest(http.MethodGet, "/api/prefectures/at?lat=35.735&lng=139.65", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}

	var body PrefectureStat
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON decode: %v (body=%s)", err, w.Body.String())
	}
	if body.Prefecture != "東京都" {
		t.Errorf("prefecture = %q, want 東京都", body.Prefecture)
	}
	if body.Count != 2 {
		t.Errorf("count = %d, want 2", body.Count)
	}
}

func TestHTTP_GetPrefectureAt_海上は404(t *testing.T) {
	r := newTestRouter(&fakeRepo{})

	req := httptest.NewRequest(http.MethodGet, "/api/prefectures/at?lat=30&lng=145", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (body=%s)", w.Code, w.Body.String())
	}
}

func TestHTTP_PostApiPins_IPハッシュをミドルウェア経由で保存する(t *testing.T) {
	const salt = "test-salt"
	repo := &fakeRepo{}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.ContextWithFallback = true // gin.Context.Value を request context へ委譲させる
	r.Use(httpmw.IPHashMiddleware(salt))
	h := NewStrictHandler(NewHandler(repo), nil)
	RegisterHandlers(r, h)

	payload := `{"nickname":"如月ファン","prefecture":"高知県","city":"高知市","municipality_code":"39201"}`
	req := httptest.NewRequest(http.MethodPost, "/api/pins", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "203.0.113.5:9999" // gin の ClientIP がこの IP を採用する
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body=%s)", w.Code, w.Body.String())
	}
	if len(repo.inserted) != 1 {
		t.Fatalf("inserted = %d件, want 1", len(repo.inserted))
	}
	// 生IPではなく salt 付きハッシュが保存されること。
	want := httpmw.HashIP(salt, "203.0.113.5")
	if got := repo.inserted[0].IPHash; got != want {
		t.Errorf("inserted IPHash = %q, want %q", got, want)
	}
}
