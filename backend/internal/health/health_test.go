package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakePinger struct{ err error }

func (f *fakePinger) Ping(ctx context.Context) error { return f.err }

type blockingPinger struct{}

func (blockingPinger) Ping(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func newTestRouter(pinger Pinger, readinessTimeout time.Duration) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewHandler(pinger, readinessTimeout).Register(r)
	return r
}

func TestHealthz_DB疎通OKなら200(t *testing.T) {
	r := newTestRouter(&fakePinger{}, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
}

func TestHealthz_DB疎通NGなら503(t *testing.T) {
	r := newTestRouter(&fakePinger{err: errors.New("boom")}, time.Second)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503 (body=%s)", w.Code, w.Body.String())
	}
}

func TestHealthz_DB疎通がタイムアウトしたら503(t *testing.T) {
	r := newTestRouter(blockingPinger{}, 10*time.Millisecond)

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
