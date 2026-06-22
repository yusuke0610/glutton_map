package geo

import "testing"

// 単純な正方形（経度0〜10, 緯度0〜10）。
func square() Polygon {
	return Polygon{Ring{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}}}
}

// 穴あき正方形（外周0〜10, 穴4〜6）。
func squareWithHole() Polygon {
	return Polygon{
		Ring{{0, 0}, {10, 0}, {10, 10}, {0, 10}, {0, 0}},
		Ring{{4, 4}, {6, 4}, {6, 6}, {4, 6}, {4, 4}},
	}
}

func TestPolygonContains(t *testing.T) {
	tests := []struct {
		name string
		poly Polygon
		p    Point
		want bool
	}{
		{"内側", square(), Point{Lng: 5, Lat: 5}, true},
		{"外側(右)", square(), Point{Lng: 15, Lat: 5}, false},
		{"外側(下)", square(), Point{Lng: 5, Lat: -1}, false},
		{"穴の中は外", squareWithHole(), Point{Lng: 5, Lat: 5}, false},
		{"穴の外の内側", squareWithHole(), Point{Lng: 1, Lat: 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.poly.Contains(tt.p); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.p, got, tt.want)
			}
		})
	}
}

func TestMultiPolygonContains(t *testing.T) {
	// 離れた2つの正方形（本土と島のイメージ）。
	mainland := square()
	island := Polygon{Ring{{20, 20}, {22, 20}, {22, 22}, {20, 22}, {20, 20}}}
	mp := MultiPolygon{mainland, island}

	if !mp.Contains(Point{Lng: 5, Lat: 5}) {
		t.Error("本土内の点が内側と判定されない")
	}
	if !mp.Contains(Point{Lng: 21, Lat: 21}) {
		t.Error("島内の点が内側と判定されない")
	}
	if mp.Contains(Point{Lng: 15, Lat: 15}) {
		t.Error("どちらの外側でもない点が内側と誤判定された")
	}
}

func TestMultiPolygonBBox(t *testing.T) {
	mp := MultiPolygon{
		square(),
		Polygon{Ring{{20, 20}, {22, 20}, {22, 22}, {20, 22}, {20, 20}}},
	}
	b := mp.BBox()
	if b.MinLng != 0 || b.MinLat != 0 || b.MaxLng != 22 || b.MaxLat != 22 {
		t.Errorf("BBox() = %+v, want {0,0,22,22}", b)
	}
}
