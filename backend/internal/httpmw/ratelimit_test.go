package httpmw

import (
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
