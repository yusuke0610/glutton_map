package pin

// Prefecture は都道府県名（OpenAPI の Prefecture enum と同一語彙）。
type Prefecture string

// Pin は地図上の1点。最小スライスでは緯度経度はただのカラム。
// Nickname/City/Comment はファン投稿で入る表示用の情報（ポップアップで見せる）。
// seed 由来のピンではこれらは空文字になる。
// IPHash は投稿者の匿名識別子（salt 付き SHA-256）。地図には出さず、提出用の
// ユニークファン集計（連投・curl の重複排除）にのみ使う。seed 由来では空文字。
type Pin struct {
	Prefecture Prefecture
	Lat        float64
	Lng        float64
	Nickname   string
	City       string
	Comment    string
	IPHash     string
}

// PinStat は提出用集計に必要な最小フィールド。地図用の取得とは別経路で使う。
type PinStat struct {
	Prefecture Prefecture
	IPHash     string
}

// Summary はヒーロー指標の集計結果。
// PrefectureCount は「何都道府県に散らばっているか」（重複を除いた数）、
// Total はピンの総数。
type Summary struct {
	PrefectureCount int
	Total           int
}

// CountByPrefecture は pins のうち target に一致するピンの件数を数える純粋関数。
// クリック地点の都道府県1件分の集計に使う（一致が無ければ 0）。
func CountByPrefecture(pins []Pin, target Prefecture) int {
	n := 0
	for _, p := range pins {
		if p.Prefecture == target {
			n++
		}
	}
	return n
}

// Summarize は pins から Summary を計算する純粋関数。
func Summarize(pins []Pin) Summary {
	seen := map[Prefecture]struct{}{}
	for _, p := range pins {
		seen[p.Prefecture] = struct{}{}
	}
	return Summary{
		PrefectureCount: len(seen),
		Total:           len(pins),
	}
}
