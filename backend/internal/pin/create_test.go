package pin

import (
	"math/rand"
	"strings"
	"testing"
)

// all47 は openapi.yaml の Prefecture enum と一致する47都道府県。
// CoordinateFor がすべての正規の都道府県で座標を返せること（重心マップの欠落バグ防止）を検証するため、
// あえて重心マップとは独立に列挙しておく。
var all47 = []Prefecture{
	"北海道", "青森県", "岩手県", "宮城県", "秋田県", "山形県", "福島県",
	"茨城県", "栃木県", "群馬県", "埼玉県", "千葉県", "東京都", "神奈川県",
	"新潟県", "富山県", "石川県", "福井県", "山梨県", "長野県", "岐阜県",
	"静岡県", "愛知県", "三重県", "滋賀県", "京都府", "大阪府", "兵庫県",
	"奈良県", "和歌山県", "鳥取県", "島根県", "岡山県", "広島県", "山口県",
	"徳島県", "香川県", "愛媛県", "高知県", "福岡県", "佐賀県", "長崎県",
	"熊本県", "大分県", "宮崎県", "鹿児島県", "沖縄県",
}

func TestCoordinateFor_全47都道府県で座標を返す(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for _, pref := range all47 {
		_, _, ok := CoordinateFor(pref, r)
		if !ok {
			t.Errorf("CoordinateFor(%q) ok=false, want true（重心マップに欠落がある）", pref)
		}
	}
}

func TestCoordinateFor_未知の県はokがfalse(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	if _, _, ok := CoordinateFor("架空県", r); ok {
		t.Error("CoordinateFor(架空県) ok=true, want false")
	}
}

func TestCoordinateFor_座標は重心付近に収まる(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	// 高知県の重心 {33.56, 133.53} 付近に収まること。何度引いても範囲内。
	const cLat, cLng, jitter = 33.56, 133.53, 0.15
	for i := 0; i < 100; i++ {
		lat, lng, ok := CoordinateFor("高知県", r)
		if !ok {
			t.Fatal("CoordinateFor(高知県) ok=false")
		}
		if lat < cLat-jitter || lat > cLat+jitter {
			t.Errorf("lat = %f, want [%f, %f]", lat, cLat-jitter, cLat+jitter)
		}
		if lng < cLng-jitter || lng > cLng+jitter {
			t.Errorf("lng = %f, want [%f, %f]", lng, cLng-jitter, cLng+jitter)
		}
	}
}

func TestPrefectureCodeFromMunicipality(t *testing.T) {
	tests := []struct {
		name string
		code string
		want string
	}{
		{name: "高知市 39201 → 39", code: "39201", want: "39"},
		{name: "札幌市中央区 01101 → 01", code: "01101", want: "01"},
		{name: "ちょうど2桁 13 → 13", code: "13", want: "13"},
		{name: "1桁は導出不能で空", code: "5", want: ""},
		{name: "空は空", code: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PrefectureCodeFromMunicipality(tt.code); got != tt.want {
				t.Errorf("PrefectureCodeFromMunicipality(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestValidateCreate(t *testing.T) {
	tests := []struct {
		name       string
		nickname   string
		prefecture string
		city       string
		comment    string
		wantErr    bool
	}{
		{name: "正常", nickname: "如月ファン", prefecture: "高知県", city: "高知市", comment: "ここの弁当が好き", wantErr: false},
		{name: "コメントは空でもよい", nickname: "ファン", prefecture: "東京都", city: "渋谷区", comment: "", wantErr: false},
		{name: "ニックネーム空はNG", nickname: "", prefecture: "高知県", city: "高知市", comment: "", wantErr: true},
		{name: "ニックネーム31文字はNG", nickname: strings.Repeat("あ", 31), prefecture: "高知県", city: "高知市", comment: "", wantErr: true},
		{name: "ニックネーム30文字はOK", nickname: strings.Repeat("あ", 30), prefecture: "高知県", city: "高知市", comment: "", wantErr: false},
		{name: "未知の都道府県はNG", nickname: "ファン", prefecture: "架空県", city: "市", comment: "", wantErr: true},
		{name: "市区町村空はNG", nickname: "ファン", prefecture: "高知県", city: "", comment: "", wantErr: true},
		{name: "市区町村51文字はNG", nickname: "ファン", prefecture: "高知県", city: strings.Repeat("市", 51), comment: "", wantErr: true},
		{name: "コメント201文字はNG", nickname: "ファン", prefecture: "高知県", city: "高知市", comment: strings.Repeat("コ", 201), wantErr: true},
		{name: "コメント200文字はOK", nickname: "ファン", prefecture: "高知県", city: "高知市", comment: strings.Repeat("コ", 200), wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreate(tt.nickname, tt.prefecture, tt.city, tt.comment)
			if tt.wantErr && err == nil {
				t.Error("err = nil, want エラー")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("err = %v, want nil", err)
			}
		})
	}
}
