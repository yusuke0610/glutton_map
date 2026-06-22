package api

import (
	"context"
	"errors"
	"testing"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// fakeRepo は PinRepository のテストダブル（フェイク）。
// DB を立てずに、返したいピンやエラーを差し込めるようにする。
type fakeRepo struct {
	pins      []pin.Pin
	err       error
	inserted  []pin.Pin // Insert で渡されたピンを記録する
	insertErr error
}

func (f *fakeRepo) GetPins(ctx context.Context) ([]pin.Pin, error) {
	return f.pins, f.err
}

// Count は handler のテストでは使わないので最小限のスタブ。
func (f *fakeRepo) Count(ctx context.Context) (int, error) { return len(f.pins), nil }
func (f *fakeRepo) Insert(ctx context.Context, p pin.Pin) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserted = append(f.inserted, p)
	return nil
}

func TestGetApiPins_集計して返す(t *testing.T) {
	// 東京2件・大阪1件 → 3都道府県ではなく2都道府県、総数3。
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "東京都", Lat: 35.6, Lng: 139.7},
		{Prefecture: "東京都", Lat: 35.7, Lng: 139.8},
		{Prefecture: "大阪府", Lat: 34.7, Lng: 135.5},
	}}
	h := NewHandler(repo)

	resp, err := h.GetApiPins(context.Background(), GetApiPinsRequestObject{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	got, ok := resp.(GetApiPins200JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型が想定外: %T", resp)
	}

	if got.PrefectureCount != 2 {
		t.Errorf("PrefectureCount = %d, want 2", got.PrefectureCount)
	}
	if got.Total != 3 {
		t.Errorf("Total = %d, want 3", got.Total)
	}
	if len(got.Pins) != 3 {
		t.Fatalf("len(Pins) = %d, want 3", len(got.Pins))
	}
	// 最小スライスでは weight は常に 1。
	for i, p := range got.Pins {
		if p.Weight == nil || *p.Weight != 1 {
			t.Errorf("Pins[%d].Weight = %v, want 1", i, p.Weight)
		}
	}
}

func strptr(s string) *string { return &s }

func TestGetApiPins_投稿フィールドを返す(t *testing.T) {
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "高知県", Lat: 33.56, Lng: 133.53, Nickname: "如月ファン", City: "高知市", Comment: "唐揚げ最高"},
	}}
	h := NewHandler(repo)

	resp, err := h.GetApiPins(context.Background(), GetApiPinsRequestObject{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	got := resp.(GetApiPins200JSONResponse)
	if len(got.Pins) != 1 {
		t.Fatalf("len(Pins) = %d, want 1", len(got.Pins))
	}
	p := got.Pins[0]
	if p.Nickname == nil || *p.Nickname != "如月ファン" {
		t.Errorf("Nickname = %v, want 如月ファン", p.Nickname)
	}
	if p.City == nil || *p.City != "高知市" {
		t.Errorf("City = %v, want 高知市", p.City)
	}
	if p.Comment == nil || *p.Comment != "唐揚げ最高" {
		t.Errorf("Comment = %v, want 唐揚げ最高", p.Comment)
	}
}

func TestPostApiPins_有効なコードで投稿が保存され201で返る(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "如月ファン",
		Prefecture:       "高知県",
		City:             "高知市",
		MunicipalityCode: "39201", // 高知市
		Comment:          strptr("ここの唐揚げ弁当が最高"),
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	created, ok := resp.(PostApiPins201JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型が想定外: %T", resp)
	}
	if created.Nickname == nil || *created.Nickname != "如月ファン" {
		t.Errorf("Nickname = %v, want 如月ファン", created.Nickname)
	}
	// リポジトリに1件保存されていること。
	if len(repo.inserted) != 1 {
		t.Fatalf("inserted = %d件, want 1", len(repo.inserted))
	}
	if repo.inserted[0].City != "高知市" {
		t.Errorf("inserted City = %q, want 高知市", repo.inserted[0].City)
	}
}

func TestPostApiPins_市区町村コード指定で境界内に生成し正規名称で保存(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "ねりまファン",
		Prefecture:       "東京都",
		City:             "ねりま", // 表記ゆれ。コード指定時は正規名称で上書きされる
		MunicipalityCode: "13120",
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	created, ok := resp.(PostApiPins201JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型が想定外: %T", resp)
	}
	// 同梱データの練馬区 bbox（経度139.560〜139.683, 緯度35.715〜35.785）内に入ること。
	if created.Lng < 139.560 || created.Lng > 139.683 {
		t.Errorf("Lng = %f, 練馬区bbox外", created.Lng)
	}
	if created.Lat < 35.715 || created.Lat > 35.785 {
		t.Errorf("Lat = %f, 練馬区bbox外", created.Lat)
	}
	// 表示用 city は正規名称（練馬区）で保存される。
	if len(repo.inserted) != 1 || repo.inserted[0].City != "練馬区" {
		t.Errorf("inserted City = %q, want 練馬区", repo.inserted[0].City)
	}
}

func TestPostApiPins_実在しないコードは400(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "ファン",
		Prefecture:       "高知県",
		City:             "高知市",
		MunicipalityCode: "00000", // 存在しないコード
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("バリデーションエラーは err ではなく 400 で返すべき: %v", err)
	}
	if _, ok := resp.(PostApiPins400JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want PostApiPins400JSONResponse", resp)
	}
	if len(repo.inserted) != 0 {
		t.Errorf("inserted = %d件, want 0", len(repo.inserted))
	}
}

func TestPostApiPins_都道府県と不一致のコードは400(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	// 北海道を選んでいるのに高知市(39201)のコード → 実在しない組み合わせなので拒否。
	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "ファン",
		Prefecture:       "北海道",
		City:             "高知市",
		MunicipalityCode: "39201", // 高知市（高知県）
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("バリデーションエラーは err ではなく 400 で返すべき: %v", err)
	}
	if _, ok := resp.(PostApiPins400JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want PostApiPins400JSONResponse", resp)
	}
	if len(repo.inserted) != 0 {
		t.Errorf("inserted = %d件, want 0", len(repo.inserted))
	}
}

func TestPostApiPins_コード未指定は400(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:   "ファン",
		Prefecture: "高知県",
		City:       "高知市",
		// MunicipalityCode 未指定（候補から選んでいない）→ 拒否。
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("バリデーションエラーは err ではなく 400 で返すべき: %v", err)
	}
	if _, ok := resp.(PostApiPins400JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want PostApiPins400JSONResponse", resp)
	}
	if len(repo.inserted) != 0 {
		t.Errorf("inserted = %d件, want 0", len(repo.inserted))
	}
}

func TestPostApiPins_不正入力は400(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:   "", // 空はNG
		Prefecture: "高知県",
		City:       "高知市",
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("バリデーションエラーは err ではなく 400 レスポンスで返すべき: %v", err)
	}
	if _, ok := resp.(PostApiPins400JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want PostApiPins400JSONResponse", resp)
	}
	// 不正入力なので保存されないこと。
	if len(repo.inserted) != 0 {
		t.Errorf("inserted = %d件, want 0", len(repo.inserted))
	}
}

func TestGetApiPins_repoエラーは型付き500を返す(t *testing.T) {
	h := NewHandler(&fakeRepo{err: errors.New("db 接続失敗")})

	resp, err := h.GetApiPins(context.Background(), GetApiPinsRequestObject{})
	// 内部エラーは Go の error ではなく、契約上の型付き 500 レスポンスで返す。
	if err != nil {
		t.Fatalf("err = %v, want nil（型付き500で返すべき）", err)
	}
	got, ok := resp.(GetApiPins500JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型 = %T, want GetApiPins500JSONResponse", resp)
	}
	if got.Message == "" {
		t.Error("Message が空。ユーザー向け文言を入れるべき")
	}
}

func TestPostApiPins_insert失敗は型付き500を返す(t *testing.T) {
	repo := &fakeRepo{insertErr: errors.New("db insert 失敗")}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname: "ファン", Prefecture: "高知県", City: "高知市", MunicipalityCode: "39201",
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("err = %v, want nil（型付き500で返すべき）", err)
	}
	if _, ok := resp.(PostApiPins500JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want PostApiPins500JSONResponse", resp)
	}
}
