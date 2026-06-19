package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

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
