package stats

import (
	"testing"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

func TestBuild_同一IPハッシュの連投はユニークファン1にまとめる(t *testing.T) {
	rows := []pin.PinStat{
		{Prefecture: "東京都", IPHash: "hashA"},
		{Prefecture: "東京都", IPHash: "hashA"}, // 連投
		{Prefecture: "東京都", IPHash: "hashA"}, // 連投
	}
	got := Build(rows)

	if got.TotalPosts != 3 {
		t.Errorf("TotalPosts = %d, want 3（投稿数の合計はそのまま）", got.TotalPosts)
	}
	if got.UniqueFans != 1 {
		t.Errorf("UniqueFans = %d, want 1（連投はユニーク化）", got.UniqueFans)
	}
	if got.PrefectureCount != 1 {
		t.Errorf("PrefectureCount = %d, want 1", got.PrefectureCount)
	}
	if got.ByPrefecture["東京都"] != 1 {
		t.Errorf("ByPrefecture[東京都] = %d, want 1", got.ByPrefecture["東京都"])
	}
}

func TestBuild_複数ファン複数県を集計する(t *testing.T) {
	rows := []pin.PinStat{
		{Prefecture: "東京都", IPHash: "hashA"},
		{Prefecture: "東京都", IPHash: "hashB"},
		{Prefecture: "東京都", IPHash: "hashA"}, // A の連投
		{Prefecture: "大阪府", IPHash: "hashC"},
	}
	got := Build(rows)

	if got.TotalPosts != 4 {
		t.Errorf("TotalPosts = %d, want 4", got.TotalPosts)
	}
	if got.UniqueFans != 3 {
		t.Errorf("UniqueFans = %d, want 3（A/B/C）", got.UniqueFans)
	}
	if got.PrefectureCount != 2 {
		t.Errorf("PrefectureCount = %d, want 2", got.PrefectureCount)
	}
	if got.ByPrefecture["東京都"] != 2 {
		t.Errorf("ByPrefecture[東京都] = %d, want 2（A/B）", got.ByPrefecture["東京都"])
	}
	if got.ByPrefecture["大阪府"] != 1 {
		t.Errorf("ByPrefecture[大阪府] = %d, want 1", got.ByPrefecture["大阪府"])
	}
}

func TestBuild_空IPハッシュのレガシー行は各1としてカウントし潰さない(t *testing.T) {
	rows := []pin.PinStat{
		{Prefecture: "高知県", IPHash: ""}, // seed/レガシー由来
		{Prefecture: "高知県", IPHash: ""}, // 別行なので別カウント
		{Prefecture: "高知県", IPHash: "hashA"},
	}
	got := Build(rows)

	if got.UniqueFans != 3 {
		t.Errorf("UniqueFans = %d, want 3（空ハッシュ2行は各1 + hashA）", got.UniqueFans)
	}
	if got.ByPrefecture["高知県"] != 3 {
		t.Errorf("ByPrefecture[高知県] = %d, want 3", got.ByPrefecture["高知県"])
	}
}

func TestBuild_空入力はゼロ値レポート(t *testing.T) {
	got := Build(nil)
	if got.TotalPosts != 0 || got.UniqueFans != 0 || got.PrefectureCount != 0 {
		t.Errorf("空入力なのに非ゼロ: %+v", got)
	}
	if got.ByPrefecture == nil {
		t.Error("ByPrefecture は空でも nil ではなく空マップであるべき（JSON で {} になる）")
	}
}
