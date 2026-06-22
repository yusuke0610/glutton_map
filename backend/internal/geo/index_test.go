package geo

import (
	"math/rand"
	"testing"
)

// テスト用の最小 GeoJSON。練馬区を模した正方形（経度139.6〜139.7, 緯度35.7〜35.8）と、
// 穴あき・MultiPolygon は polygon_test.go で別途検証済みなのでここは単純形でよい。
const sampleGeoJSON = `{
  "type": "FeatureCollection",
  "features": [
    {
      "type": "Feature",
      "properties": {"code": "13120", "prefecture": "東京都", "name": "練馬区", "kana": "ねりまく"},
      "geometry": {
        "type": "Polygon",
        "coordinates": [[[139.6,35.7],[139.7,35.7],[139.7,35.8],[139.6,35.8],[139.6,35.7]]]
      }
    },
    {
      "type": "Feature",
      "properties": {"code": "39201", "prefecture": "高知県", "name": "高知市", "kana": "こうちし"},
      "geometry": {
        "type": "MultiPolygon",
        "coordinates": [[[[133.5,33.5],[133.6,33.5],[133.6,33.6],[133.5,33.6],[133.5,33.5]]]]
      }
    }
  ]
}`

func mustParse(t *testing.T) *Index {
	t.Helper()
	ix, err := ParseGeoJSON([]byte(sampleGeoJSON))
	if err != nil {
		t.Fatalf("ParseGeoJSON: %v", err)
	}
	return ix
}

func TestParseGeoJSON(t *testing.T) {
	ix := mustParse(t)
	if ix.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", ix.Len())
	}
	m, ok := ix.Get("13120")
	if !ok {
		t.Fatal("13120 が見つからない")
	}
	if m.Name != "練馬区" || m.Prefecture != "東京都" {
		t.Errorf("Get(13120) = %q/%q, want 練馬区/東京都", m.Prefecture, m.Name)
	}
}

func TestSamplePointIsInside(t *testing.T) {
	ix := mustParse(t)
	r := rand.New(rand.NewSource(1))
	m, _ := ix.Get("13120")
	for i := 0; i < 200; i++ {
		lat, lng, ok := ix.SamplePoint("13120", r)
		if !ok {
			t.Fatal("SamplePoint が ok=false")
		}
		if !m.Geometry.Contains(Point{Lng: lng, Lat: lat}) {
			t.Fatalf("生成点(%f,%f)が境界の外", lat, lng)
		}
	}
}

func TestSamplePointDeterministic(t *testing.T) {
	ix := mustParse(t)
	lat1, lng1, _ := ix.SamplePoint("13120", rand.New(rand.NewSource(42)))
	lat2, lng2, _ := ix.SamplePoint("13120", rand.New(rand.NewSource(42)))
	if lat1 != lat2 || lng1 != lng2 {
		t.Errorf("同一シードで結果が異なる: (%f,%f) vs (%f,%f)", lat1, lng1, lat2, lng2)
	}
}

func TestSamplePointUnknownCode(t *testing.T) {
	ix := mustParse(t)
	if _, _, ok := ix.SamplePoint("99999", rand.New(rand.NewSource(1))); ok {
		t.Error("未知コードで ok=true になった")
	}
}

func TestParseGeoJSONRejectsRepOutside(t *testing.T) {
	// 共線・面積ゼロの退化ポリゴンは内部点を持たず、フォールバックの代表点も境界外になる。
	// このとき SamplePoint が境界外の点を ok=true で返さないよう、パース時点で弾く。
	const degenerate = `{
	  "type": "FeatureCollection",
	  "features": [
	    {
	      "type": "Feature",
	      "properties": {"code": "00000", "prefecture": "X", "name": "退化"},
	      "geometry": {"type": "Polygon", "coordinates": [[[0,0],[1,1],[2,2],[0,0]]]}
	    }
	  ]
	}`
	if _, err := ParseGeoJSON([]byte(degenerate)); err == nil {
		t.Fatal("代表点が境界外なのにエラーにならなかった")
	}
}

func TestRepresentativePointInside(t *testing.T) {
	ix := mustParse(t)
	for _, code := range []string{"13120", "39201"} {
		m, _ := ix.Get(code)
		if !m.Geometry.Contains(m.Rep) {
			t.Errorf("%s の代表点 %+v が境界の外", code, m.Rep)
		}
	}
}
