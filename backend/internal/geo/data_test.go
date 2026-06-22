package geo

import (
	"math/rand"
	"testing"
)

// 同梱データがロードでき、代表例（練馬区）が境界内に点を生成できることを確認する。
func TestDefaultLoads(t *testing.T) {
	ix, err := Default()
	if err != nil {
		t.Fatalf("Default: %v", err)
	}
	if ix.Len() < 1 {
		t.Fatalf("Len() = %d, want >= 1", ix.Len())
	}
	m, ok := ix.Get("13120")
	if !ok {
		t.Fatal("練馬区(13120)が同梱データに無い")
	}
	lat, lng, ok := ix.SamplePoint("13120", rand.New(rand.NewSource(1)))
	if !ok || !m.Geometry.Contains(Point{Lng: lng, Lat: lat}) {
		t.Fatalf("練馬区の生成点(%f,%f)が境界外", lat, lng)
	}
}
