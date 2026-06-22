package geo

import (
	_ "embed"
	"sync"
)

// embeddedGeoJSON はビルド時に同梱する全国市区町村の境界データ。
// `make gen-municipalities` が data/municipalities.geojson を生成・上書きする。
//
//go:embed data/municipalities.geojson
var embeddedGeoJSON []byte

var (
	defaultOnce  sync.Once
	defaultIndex *Index
	defaultErr   error
)

// Default は同梱データから構築した索引を返す（初回のみ解析、以降キャッシュ）。
// 同梱データはビルド時に固定なので、解析失敗は実質バグ。エラーは error で返す。
func Default() (*Index, error) {
	defaultOnce.Do(func() {
		defaultIndex, defaultErr = ParseGeoJSON(embeddedGeoJSON)
	})
	return defaultIndex, defaultErr
}
