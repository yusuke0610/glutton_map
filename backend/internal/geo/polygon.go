// Package geo は行政区域ポリゴンを保持し、点が境界内かを判定する（純 Go / cgo なし）。
// DB には一切触れず、市区町村の座標生成のための静的な幾何処理だけを担う。
package geo

// Point は座標1点。GeoJSON に合わせて経度(Lng)・緯度(Lat)で持つ。
type Point struct {
	Lng float64
	Lat float64
}

// Ring は閉じた多角形の輪（最初と最後の点が同じ）。
type Ring []Point

// Polygon は1つの面。[0] が外周、[1:] が穴（くり抜き）。
type Polygon []Ring

// MultiPolygon は複数の面（本土＋島など）の集合。
type MultiPolygon []Polygon

// BBox は緯度経度の最小・最大で表す外接矩形。
type BBox struct {
	MinLng float64
	MinLat float64
	MaxLng float64
	MaxLat float64
}

// Contains は点が外接矩形の内側（境界含む）かを返す。点内包判定の安価な足切りに使う。
func (b BBox) Contains(p Point) bool {
	return p.Lng >= b.MinLng && p.Lng <= b.MaxLng && p.Lat >= b.MinLat && p.Lat <= b.MaxLat
}

// ringContains はレイキャスティング法で点が輪の内側かを判定する。
// 境界線上の点の扱いは未定義（用途上、内外どちらでも害がない）。
func ringContains(r Ring, p Point) bool {
	inside := false
	n := len(r)
	if n < 3 {
		return false
	}
	j := n - 1
	for i := 0; i < n; i++ {
		// 点の緯度が辺の緯度区間をまたぐとき、辺との交点の経度を求めて左右を数える。
		if (r[i].Lat > p.Lat) != (r[j].Lat > p.Lat) {
			xInt := (r[j].Lng-r[i].Lng)*(p.Lat-r[i].Lat)/(r[j].Lat-r[i].Lat) + r[i].Lng
			if p.Lng < xInt {
				inside = !inside
			}
		}
		j = i
	}
	return inside
}

// Contains は外周の内側かつどの穴にも入っていなければ true。
func (poly Polygon) Contains(p Point) bool {
	if len(poly) == 0 || !ringContains(poly[0], p) {
		return false
	}
	for _, hole := range poly[1:] {
		if ringContains(hole, p) {
			return false
		}
	}
	return true
}

// Contains はいずれかの面に含まれれば true。
func (mp MultiPolygon) Contains(p Point) bool {
	for _, poly := range mp {
		if poly.Contains(p) {
			return true
		}
	}
	return false
}

// BBox は全ての点を囲む外接矩形を返す。空のときゼロ値。
func (mp MultiPolygon) BBox() BBox {
	first := true
	var b BBox
	for _, poly := range mp {
		for _, ring := range poly {
			for _, pt := range ring {
				if first {
					b = BBox{pt.Lng, pt.Lat, pt.Lng, pt.Lat}
					first = false
					continue
				}
				if pt.Lng < b.MinLng {
					b.MinLng = pt.Lng
				}
				if pt.Lng > b.MaxLng {
					b.MaxLng = pt.Lng
				}
				if pt.Lat < b.MinLat {
					b.MinLat = pt.Lat
				}
				if pt.Lat > b.MaxLat {
					b.MaxLat = pt.Lat
				}
			}
		}
	}
	return b
}
