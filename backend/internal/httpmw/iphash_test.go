package httpmw

import (
	"context"
	"testing"
)

func TestHashIP_同一saltと同一IPは同じハッシュ(t *testing.T) {
	a := HashIP("salt", "1.1.1.1")
	b := HashIP("salt", "1.1.1.1")
	if a != b {
		t.Errorf("同じ salt/IP なのにハッシュが違う: %q != %q", a, b)
	}
	// 生IPがそのまま残らない（一方向ハッシュであること）。
	if a == "1.1.1.1" || a == "" {
		t.Errorf("ハッシュが生IPまたは空: %q", a)
	}
}

func TestHashIP_saltが違えばハッシュも違う(t *testing.T) {
	if HashIP("salt-A", "1.1.1.1") == HashIP("salt-B", "1.1.1.1") {
		t.Error("salt が違えばハッシュも変わるべき")
	}
}

func TestHashIP_IPが違えばハッシュも違う(t *testing.T) {
	if HashIP("salt", "1.1.1.1") == HashIP("salt", "2.2.2.2") {
		t.Error("IP が違えばハッシュも変わるべき")
	}
}

func TestIPHashFrom_contextに載せた値を取り出せる(t *testing.T) {
	ctx := withIPHash(context.Background(), "deadbeef")
	if got := IPHashFrom(ctx); got != "deadbeef" {
		t.Errorf("IPHashFrom = %q, want deadbeef", got)
	}
}

func TestIPHashFrom_未設定なら空文字(t *testing.T) {
	if got := IPHashFrom(context.Background()); got != "" {
		t.Errorf("IPHashFrom = %q, want 空文字", got)
	}
}
