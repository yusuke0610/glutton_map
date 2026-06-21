package httpmw

import (
	"fmt"
	"testing"
	"time"
)

func TestLimiter_同一IPの連投はクールダウン中拒否する(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(3 * time.Second)
	l.now = func() time.Time { return now }

	if !l.Allow("1.1.1.1") {
		t.Fatal("1回目は許可されるべき")
	}
	if l.Allow("1.1.1.1") {
		t.Error("クールダウン中の2回目は拒否されるべき")
	}

	// クールダウンを過ぎれば再び許可。
	now = now.Add(3 * time.Second)
	if !l.Allow("1.1.1.1") {
		t.Error("クールダウン経過後は許可されるべき")
	}
}

func TestLimiter_クールダウン済みエントリは破棄されメモリが膨張しない(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(3 * time.Second)
	l.now = func() time.Time { return now }

	// 多数の使い捨て IP がアクセスする（ローテーションするスパムを想定）。
	for i := 0; i < 100; i++ {
		l.Allow(fmt.Sprintf("10.0.0.%d", i))
	}
	// クールダウンを過ぎてから別 IP がアクセスすると、古いエントリは破棄される。
	now = now.Add(3 * time.Second)
	l.Allow("1.1.1.1")

	if got := len(l.last); got > 1 {
		t.Errorf("len(last) = %d, want 1（クールダウン済みは破棄されるべき）", got)
	}
}

func TestLimiter_別IPは互いに干渉しない(t *testing.T) {
	now := time.Unix(0, 0)
	l := NewLimiter(3 * time.Second)
	l.now = func() time.Time { return now }

	if !l.Allow("1.1.1.1") {
		t.Fatal("IP-A 1回目は許可されるべき")
	}
	if !l.Allow("2.2.2.2") {
		t.Error("IP-B は別IPなので許可されるべき")
	}
}
