package geo

import (
	"encoding/json"
	"fmt"
	"math/rand"
)

// sampleMaxTries は棄却サンプリングの試行上限。これを超えたら代表点へフォールバックする。
const sampleMaxTries = 60

// Municipality は1市区町村の幾何情報。
type Municipality struct {
	Code       string
	Prefecture string
	Name       string
	Kana       string
	Geometry   MultiPolygon
	Rep        Point // 必ず境界内に入る代表点（フォールバック用）
	bbox       BBox
}

// Index は全国地方公共団体コード→市区町村の索引。
type Index struct {
	byCode map[string]*Municipality
}

// Len は登録件数を返す。
func (ix *Index) Len() int { return len(ix.byCode) }

// Get はコードで市区町村を引く。
func (ix *Index) Get(code string) (*Municipality, bool) {
	m, ok := ix.byCode[code]
	return m, ok
}

// SamplePoint はコードの境界内の点を返す（lat, lng）。
// bbox 内の棄却サンプリングで内側の点を探し、見つからなければ代表点を返す（必ず内側）。
func (ix *Index) SamplePoint(code string, r *rand.Rand) (lat, lng float64, ok bool) {
	m, found := ix.byCode[code]
	if !found {
		return 0, 0, false
	}
	b := m.bbox
	for i := 0; i < sampleMaxTries; i++ {
		p := Point{
			Lng: b.MinLng + r.Float64()*(b.MaxLng-b.MinLng),
			Lat: b.MinLat + r.Float64()*(b.MaxLat-b.MinLat),
		}
		if m.Geometry.Contains(p) {
			return p.Lat, p.Lng, true
		}
	}
	return m.Rep.Lat, m.Rep.Lng, true
}

// --- GeoJSON パース ---

type featureCollection struct {
	Features []feature `json:"features"`
}

type feature struct {
	Properties struct {
		Code       string  `json:"code"`
		Prefecture string  `json:"prefecture"`
		Name       string  `json:"name"`
		Kana       string  `json:"kana"`
		RepLng     float64 `json:"rep_lng"`
		RepLat     float64 `json:"rep_lat"`
	} `json:"properties"`
	Geometry geometry `json:"geometry"`
}

type geometry struct {
	Type        string          `json:"type"`
	Coordinates json.RawMessage `json:"coordinates"`
}

// ParseGeoJSON は FeatureCollection を読み、コード索引を構築する。
// 各市区町村の代表点(Rep)はロード時に計算し、必ず境界内になるようにする。
func ParseGeoJSON(data []byte) (*Index, error) {
	var fc featureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("geojson 解析: %w", err)
	}
	ix := &Index{byCode: make(map[string]*Municipality, len(fc.Features))}
	for _, f := range fc.Features {
		if f.Properties.Code == "" {
			continue
		}
		mp, err := decodeGeometry(f.Geometry)
		if err != nil {
			return nil, fmt.Errorf("%s の geometry: %w", f.Properties.Code, err)
		}
		m := &Municipality{
			Code:       f.Properties.Code,
			Prefecture: f.Properties.Prefecture,
			Name:       f.Properties.Name,
			Kana:       f.Properties.Kana,
			Geometry:   mp,
			bbox:       mp.BBox(),
		}
		// 代表点はパイプラインが事前計算した値（rep_lng/rep_lat）を優先する。
		// 全国規模ではロード時のグリッド走査は重いため、事前計算が無い／境界外のときだけ走査でフォールバックする。
		rep := Point{Lng: f.Properties.RepLng, Lat: f.Properties.RepLat}
		if (rep.Lng != 0 || rep.Lat != 0) && mp.Contains(rep) {
			m.Rep = rep
		} else {
			m.Rep = representativePoint(mp, m.bbox)
		}
		// 代表点が境界外（フォールバックのグリッド走査が内部点を取り逃した／退化ポリゴン等）なら、
		// SamplePoint が境界外の点を返さないようデータ不正として弾く。
		if !mp.Contains(m.Rep) {
			return nil, fmt.Errorf("%s の代表点が境界外です", f.Properties.Code)
		}
		ix.byCode[m.Code] = m
	}
	return ix, nil
}

// decodeGeometry は Polygon / MultiPolygon を MultiPolygon に正規化する。
func decodeGeometry(g geometry) (MultiPolygon, error) {
	switch g.Type {
	case "Polygon":
		var coords [][][2]float64
		if err := json.Unmarshal(g.Coordinates, &coords); err != nil {
			return nil, err
		}
		return MultiPolygon{toPolygon(coords)}, nil
	case "MultiPolygon":
		var coords [][][][2]float64
		if err := json.Unmarshal(g.Coordinates, &coords); err != nil {
			return nil, err
		}
		mp := make(MultiPolygon, 0, len(coords))
		for _, poly := range coords {
			mp = append(mp, toPolygon(poly))
		}
		return mp, nil
	default:
		return nil, fmt.Errorf("未対応の geometry type: %q", g.Type)
	}
}

func toPolygon(rings [][][2]float64) Polygon {
	poly := make(Polygon, 0, len(rings))
	for _, ring := range rings {
		r := make(Ring, 0, len(ring))
		for _, c := range ring {
			r = append(r, Point{Lng: c[0], Lat: c[1]})
		}
		poly = append(poly, r)
	}
	return poly
}

// representativePoint は bbox 上を粗いグリッドで走査し、境界内かつ中心に近い点を返す。
// ポリゴンに面積がある限り内側の点が必ず1つは見つかる。
func representativePoint(mp MultiPolygon, b BBox) Point {
	const grid = 64
	cx := (b.MinLng + b.MaxLng) / 2
	cy := (b.MinLat + b.MaxLat) / 2
	best := Point{Lng: cx, Lat: cy}
	bestDist := -1.0
	for i := 1; i < grid; i++ {
		for j := 1; j < grid; j++ {
			p := Point{
				Lng: b.MinLng + (b.MaxLng-b.MinLng)*float64(i)/grid,
				Lat: b.MinLat + (b.MaxLat-b.MinLat)*float64(j)/grid,
			}
			if !mp.Contains(p) {
				continue
			}
			dx, dy := p.Lng-cx, p.Lat-cy
			d := dx*dx + dy*dy
			if bestDist < 0 || d < bestDist {
				bestDist = d
				best = p
			}
		}
	}
	return best
}
