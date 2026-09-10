package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type fakePinger struct{ err error }

func (f *fakePinger) Ping(ctx context.Context) error { return f.err }

func newTestRouter(pinger Pinger) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(pinger).Register(r)
	return r
}

func TestHealthz_DB疎通OKなら200(t *testing.T) {
	r := newTestRouter(&fakePinger{})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
}

func TestHealthz_DB疎通NGなら503(t *testing.T) {
	r := newTestRouter(&fakePinger{err: errors.New("boom")})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body=%s)", w.Code, w.Body.String())
	}
}

func TestCheck_200なら成功(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if err := Check(context.Background(), ts.Client(), ts.URL); err != nil {
		t.Errorf("Check: %v", err)
	}
}

func TestCheck_200以外はエラー(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	if err := Check(context.Background(), ts.Client(), ts.URL); err == nil {
		t.Error("503 なのに Check がエラーを返さなかった")
	}
}
