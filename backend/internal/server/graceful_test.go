package server

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/kisaragi-ai-map/backend/internal/logger"
)

func TestRun_進行中のリクエストを完了させてから終了する(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- Run(ctx, srv, ln, 5*time.Second, logger.New(io.Discard))
	}()

	respCh := make(chan *http.Response, 1)
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err != nil {
			t.Errorf("http.Get: %v", err)
			return
		}
		respCh <- resp
	}()

	<-started       // ハンドラがリクエストを受け付けたのを確認してから
	cancel()        // SIGTERM 相当でシャットダウンを開始させ
	close(release)  // その後にハンドラの処理を進める（進行中のリクエストとして扱われる）

	select {
	case resp := <-respCh:
		if resp.StatusCode != http.StatusOK {
			t.Errorf("status = %d, want 200", resp.StatusCode)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("進行中のリクエストが完了する前にタイムアウトした")
	}

	if err := <-runErrCh; err != nil {
		t.Errorf("Run: %v", err)
	}
}

func TestRun_猶予を過ぎたら強制終了して無限待ちしない(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release // このテストでは解放しない = ハンドラが猶予内に終わらない状況を模す
		w.WriteHeader(http.StatusOK)
	})}
	defer close(release) // テスト終了時にリークさせない

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- Run(ctx, srv, ln, 50*time.Millisecond, logger.New(io.Discard))
	}()

	go func() {
		//nolint:errcheck // タイムアウトで接続が切られる想定のため戻り値は見ない
		http.Get("http://" + ln.Addr().String())
	}()
	<-started
	cancel()

	select {
	case err := <-runErrCh:
		if err == nil {
			t.Error("猶予超過なのに Run がエラーを返さなかった")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("猶予を過ぎても Run が戻ってこない（無限待ちしている）")
	}
}
