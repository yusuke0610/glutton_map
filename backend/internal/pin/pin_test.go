package pin

import "testing"

func TestSummarize(t *testing.T) {
	tests := []struct {
		name                string
		pins                []Pin
		wantPrefectureCount int
		wantTotal           int
		wantUniqueFans      int
	}{
		{
			name:                "空のスライスは0件0都道府県0ファン",
			pins:                nil,
			wantPrefectureCount: 0,
			wantTotal:           0,
			wantUniqueFans:      0,
		},
		{
			name:                "1件なら1都道府県1ファン",
			pins:                []Pin{{Prefecture: "東京都", IPHash: "h1"}},
			wantPrefectureCount: 1,
			wantTotal:           1,
			wantUniqueFans:      1,
		},
		{
			name:                "同じ都道府県が2件でも都道府県数は1（重複除去）",
			pins:                []Pin{{Prefecture: "東京都", IPHash: "h1"}, {Prefecture: "東京都", IPHash: "h2"}},
			wantPrefectureCount: 1,
			wantTotal:           2,
			wantUniqueFans:      2,
		},
		{
			name:                "異なる都道府県は別々に数える",
			pins:                []Pin{{Prefecture: "東京都", IPHash: "h1"}, {Prefecture: "大阪府", IPHash: "h2"}, {Prefecture: "東京都", IPHash: "h1"}},
			wantPrefectureCount: 2,
			wantTotal:           3,
			wantUniqueFans:      2,
		},
		{
			name: "同一ip_hashの連投はユニークファン数を増やさない（totalとprefecture_countは変わらない）",
			pins: []Pin{
				{Prefecture: "高知県", IPHash: "h1"},
				{Prefecture: "高知県", IPHash: "h1"},
				{Prefecture: "高知県", IPHash: "h1"},
			},
			wantPrefectureCount: 1,
			wantTotal:           3,
			wantUniqueFans:      1,
		},
		{
			name: "ip_hashが空(seed由来)の行は畳まず各行を1ファンとして数える",
			pins: []Pin{
				{Prefecture: "東京都", IPHash: ""},
				{Prefecture: "東京都", IPHash: ""},
			},
			wantPrefectureCount: 1,
			wantTotal:           2,
			wantUniqueFans:      2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Summarize(tt.pins)

			if got.PrefectureCount != tt.wantPrefectureCount {
				t.Errorf("PrefectureCount = %d, want %d", got.PrefectureCount, tt.wantPrefectureCount)
			}
			if got.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", got.Total, tt.wantTotal)
			}
			if got.UniqueFans != tt.wantUniqueFans {
				t.Errorf("UniqueFans = %d, want %d", got.UniqueFans, tt.wantUniqueFans)
			}
		})
	}
}

func TestCountByPrefecture(t *testing.T) {
	pins := []Pin{
		{Prefecture: "東京都"},
		{Prefecture: "大阪府"},
		{Prefecture: "東京都"},
	}
	tests := []struct {
		name   string
		target Prefecture
		want   int
	}{
		{"対象県の件数を数える", "東京都", 2},
		{"別の県は別に数える", "大阪府", 1},
		{"ピンが無い県は0件", "高知県", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountByPrefecture(pins, tt.target); got != tt.want {
				t.Errorf("CountByPrefecture(%q) = %d, want %d", tt.target, got, tt.want)
			}
		})
	}
}
