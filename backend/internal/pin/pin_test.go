package pin

import "testing"

func TestSummarize(t *testing.T) {
	tests := []struct {
		name                string
		pins                []Pin
		wantPrefectureCount int
		wantTotal           int
	}{
		{
			name:                "空のスライスは0件0都道府県",
			pins:                nil,
			wantPrefectureCount: 0,
			wantTotal:           0,
		},
		{
			name:                "1件なら1都道府県",
			pins:                []Pin{{Prefecture: "東京都"}},
			wantPrefectureCount: 1,
			wantTotal:           1,
		},
		{
			name:                "同じ都道府県が2件でも都道府県数は1（重複除去）",
			pins:                []Pin{{Prefecture: "東京都"}, {Prefecture: "東京都"}},
			wantPrefectureCount: 1,
			wantTotal:           2,
		},
		{
			name:                "異なる都道府県は別々に数える",
			pins:                []Pin{{Prefecture: "東京都"}, {Prefecture: "大阪府"}, {Prefecture: "東京都"}},
			wantPrefectureCount: 2,
			wantTotal:           3,
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
		})
	}
}
