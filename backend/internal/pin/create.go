package pin

import (
	"fmt"
	"math/rand"
	"unicode/utf8"
)

// 入力フィールドの文字数上限。openapi.yaml の maxLength と一致させること。
const (
	maxNicknameLen = 30
	maxCityLen     = 50
	maxCommentLen  = 200
)

// jitterRange はピンの座標を都道府県重心から散らすゆらぎ幅（±jitterRange/2）。
const jitterRange = 0.3

// prefectureCentroids は各都道府県の代表座標 {lat, lng}。
// 投稿ピンの座標はここから ±0.15 のゆらぎで生成し、正確な現在地は受け取らない（プライバシー）。
// openapi.yaml の Prefecture enum 47件すべてを網羅すること。
var prefectureCentroids = map[Prefecture][2]float64{
	"北海道": {43.06, 141.35}, "青森県": {40.82, 140.74}, "岩手県": {39.70, 141.15},
	"宮城県": {38.27, 140.87}, "秋田県": {39.72, 140.10}, "山形県": {38.24, 140.36},
	"福島県": {37.75, 140.47}, "茨城県": {36.34, 140.45}, "栃木県": {36.57, 139.88},
	"群馬県": {36.39, 139.06}, "埼玉県": {35.86, 139.65}, "千葉県": {35.61, 140.12},
	"東京都": {35.69, 139.69}, "神奈川県": {35.45, 139.64}, "新潟県": {37.90, 139.02},
	"富山県": {36.70, 137.21}, "石川県": {36.59, 136.63}, "福井県": {36.07, 136.22},
	"山梨県": {35.66, 138.57}, "長野県": {36.65, 138.18}, "岐阜県": {35.39, 136.72},
	"静岡県": {34.98, 138.38}, "愛知県": {35.18, 136.91}, "三重県": {34.73, 136.51},
	"滋賀県": {35.00, 135.87}, "京都府": {35.02, 135.76}, "大阪府": {34.69, 135.50},
	"兵庫県": {34.69, 134.04}, "奈良県": {34.69, 135.83}, "和歌山県": {34.23, 135.17},
	"鳥取県": {35.50, 134.24}, "島根県": {35.47, 133.05}, "岡山県": {34.66, 133.93},
	"広島県": {34.40, 132.46}, "山口県": {34.19, 131.47}, "徳島県": {34.07, 134.56},
	"香川県": {34.34, 134.04}, "愛媛県": {33.84, 132.77}, "高知県": {33.56, 133.53},
	"福岡県": {33.61, 130.42}, "佐賀県": {33.25, 130.30}, "長崎県": {32.74, 129.87},
	"熊本県": {32.79, 130.74}, "大分県": {33.24, 131.61}, "宮崎県": {31.91, 131.42},
	"鹿児島県": {31.56, 130.56}, "沖縄県": {26.21, 127.68},
}

// CoordinateFor は都道府県の重心に ±jitterRange/2 のゆらぎを足した座標を返す。
// 未知の都道府県のとき ok=false。乱数源を引数で受け取りテスト可能にする。
func CoordinateFor(pref Prefecture, r *rand.Rand) (lat, lng float64, ok bool) {
	c, ok := prefectureCentroids[pref]
	if !ok {
		return 0, 0, false
	}
	jitter := func() float64 { return r.Float64()*jitterRange - jitterRange/2 }
	return c[0] + jitter(), c[1] + jitter(), true
}

// PrefectureCodeFromMunicipality は全国地方公共団体コード（市区町村コード）の先頭2桁を
// 都道府県コード（JIS X 0401）として返す。座標は municipality_code から境界内に生成済みなので、
// 都道府県コードは point-in-polygon で再判定せずコード先頭2桁から導出すれば完全一致する。
// 2桁に満たないコードは導出不能として空文字を返す。
func PrefectureCodeFromMunicipality(code string) string {
	if len(code) < 2 {
		return ""
	}
	return code[:2]
}

// EachPrefecture は全都道府県とその重心を fn に渡す（seed 用）。
// 重心マップを外へ漏らさずに走査できるようにする。
func EachPrefecture(fn func(pref Prefecture)) {
	for pref := range prefectureCentroids {
		fn(pref)
	}
}

// ValidateCreate はファン投稿の入力を検証する純粋関数。
// 文字数は UTF-8 のルーン数（見た目の文字数）で数える。
func ValidateCreate(nickname, prefecture, city, comment string) error {
	if n := utf8.RuneCountInString(nickname); n < 1 || n > maxNicknameLen {
		return fmt.Errorf("ニックネームは1〜%d文字で入力してください", maxNicknameLen)
	}
	if _, ok := prefectureCentroids[Prefecture(prefecture)]; !ok {
		return fmt.Errorf("都道府県が不正です: %q", prefecture)
	}
	if n := utf8.RuneCountInString(city); n < 1 || n > maxCityLen {
		return fmt.Errorf("市区町村は1〜%d文字で入力してください", maxCityLen)
	}
	if n := utf8.RuneCountInString(comment); n > maxCommentLen {
		return fmt.Errorf("コメントは%d文字以内で入力してください", maxCommentLen)
	}
	return nil
}
