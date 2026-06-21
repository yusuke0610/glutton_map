package pin

// Prefecture は都道府県名（OpenAPI の Prefecture enum と同一語彙）。
type Prefecture string

// Pin は地図上の1点。最小スライスでは緯度経度はただのカラム。
// Nickname/City/Comment はファン投稿で入る表示用の情報（ポップアップで見せる）。
// seed 由来のピンではこれらは空文字になる。
type Pin struct {
	Prefecture Prefecture
	Lat        float64
	Lng        float64
	Nickname   string
	City       string
	Comment    string
}

// Summary はヒーロー指標の集計結果。
// PrefectureCount は「何都道府県に散らばっているか」（重複を除いた数）、
// Total はピンの総数。
type Summary struct {
	PrefectureCount int
	Total           int
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
