package api

import (
	"context"
	"errors"
	"math/rand"
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

// ListForStats は handler のテストでは使わないので最小限のスタブ。
func (f *fakeRepo) ListForStats(ctx context.Context) ([]pin.PinStat, error) { return nil, nil }

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

func TestPostApiPins_流入元と匿名トークンと都道府県コードを保存する(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "如月ファン",
		Prefecture:       "高知県",
		City:             "高知市",
		MunicipalityCode: "39201", // 高知市 → 都道府県コードは先頭2桁の "39"
		AnonToken:        strptr("anon-xyz"),
		UtmSource:        strptr("twitter"),
		UtmMedium:        strptr("social"),
		UtmCampaign:      strptr("fan_share"),
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if _, ok := resp.(PostApiPins201JSONResponse); !ok {
		t.Fatalf("レスポンス型が想定外: %T", resp)
	}
	if len(repo.inserted) != 1 {
		t.Fatalf("inserted = %d件, want 1", len(repo.inserted))
	}
	got := repo.inserted[0]
	// 都道府県コードは municipality_code の先頭2桁から導出する（point-in-polygon は使わない）。
	if got.PrefectureCode != "39" {
		t.Errorf("PrefectureCode = %q, want 39", got.PrefectureCode)
	}
	if got.AnonToken != "anon-xyz" {
		t.Errorf("AnonToken = %q, want anon-xyz", got.AnonToken)
	}
	if got.UTMSource != "twitter" || got.UTMMedium != "social" || got.UTMCampaign != "fan_share" {
		t.Errorf("UTM = %q/%q/%q, want twitter/social/fan_share", got.UTMSource, got.UTMMedium, got.UTMCampaign)
	}
	// 分析用フィールドは地図用レスポンスには出さない（PII 配慮・payload 軽量化）。
	created := resp.(PostApiPins201JSONResponse)
	_ = created
}

func TestPostApiPins_流入元は任意で未指定でも201(t *testing.T) {
	repo := &fakeRepo{}
	h := NewHandler(repo)

	// utm/anon_token を一切付けない最初のピン。認証も計測値も無くても通ること（転換率優先）。
	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:         "ファン",
		Prefecture:       "高知県",
		City:             "高知市",
		MunicipalityCode: "39201",
	}}
	resp, err := h.PostApiPins(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if _, ok := resp.(PostApiPins201JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want 201", resp)
	}
	got := repo.inserted[0]
	if got.AnonToken != "" || got.UTMSource != "" {
		t.Errorf("未指定なら空のはず: AnonToken=%q UTMSource=%q", got.AnonToken, got.UTMSource)
	}
	// コードがあるので都道府県コードは導出される。
	if got.PrefectureCode != "39" {
		t.Errorf("PrefectureCode = %q, want 39", got.PrefectureCode)
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

func TestPostApiPins_縮退モードでもコード未指定は400(t *testing.T) {
	repo := &fakeRepo{}
	// 境界データのロードに失敗した縮退モード（muni == nil）を再現する。
	h := &Handler{repo: repo, rng: rand.New(rand.NewSource(1))}

	req := PostApiPinsRequestObject{Body: &PostApiPinsJSONRequestBody{
		Nickname:   "ファン",
		Prefecture: "高知県",
		City:       "高知市",
		// MunicipalityCode 未指定。縮退モードでも必須契約を守り 201 で通さない。
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

func TestGetPrefectureAt_県内なら件数を返す(t *testing.T) {
	repo := &fakeRepo{pins: []pin.Pin{
		{Prefecture: "東京都", Lat: 35.6, Lng: 139.7},
		{Prefecture: "東京都", Lat: 35.7, Lng: 139.8},
		{Prefecture: "大阪府", Lat: 34.7, Lng: 135.5},
	}}
	h := NewHandler(repo)

	// 練馬区あたり（東京都）の座標。
	req := GetPrefectureAtRequestObject{Params: GetPrefectureAtParams{Lat: 35.735, Lng: 139.65}}
	resp, err := h.GetPrefectureAt(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	got, ok := resp.(GetPrefectureAt200JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型 = %T, want GetPrefectureAt200JSONResponse", resp)
	}
	if got.Prefecture != "東京都" {
		t.Errorf("Prefecture = %q, want 東京都", got.Prefecture)
	}
	if got.Count != 2 {
		t.Errorf("Count = %d, want 2", got.Count)
	}
}

func TestGetPrefectureAt_ピンが無い県は0件で200(t *testing.T) {
	repo := &fakeRepo{pins: []pin.Pin{{Prefecture: "大阪府", Lat: 34.7, Lng: 135.5}}}
	h := NewHandler(repo)

	req := GetPrefectureAtRequestObject{Params: GetPrefectureAtParams{Lat: 35.735, Lng: 139.65}}
	resp, err := h.GetPrefectureAt(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	got, ok := resp.(GetPrefectureAt200JSONResponse)
	if !ok {
		t.Fatalf("レスポンス型 = %T, want GetPrefectureAt200JSONResponse", resp)
	}
	if got.Prefecture != "東京都" || got.Count != 0 {
		t.Errorf("got %q/%d, want 東京都/0", got.Prefecture, got.Count)
	}
}

func TestGetPrefectureAt_海上は404(t *testing.T) {
	h := NewHandler(&fakeRepo{})

	// 太平洋上（どの市区町村にも属さない）。
	req := GetPrefectureAtRequestObject{Params: GetPrefectureAtParams{Lat: 30.0, Lng: 145.0}}
	resp, err := h.GetPrefectureAt(context.Background(), req)
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if _, ok := resp.(GetPrefectureAt404JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want GetPrefectureAt404JSONResponse", resp)
	}
}

func TestGetPrefectureAt_縮退モードは500(t *testing.T) {
	// 境界データのロードに失敗した縮退モード（muni == nil）では判定不能なので 500。
	h := &Handler{repo: &fakeRepo{}, rng: rand.New(rand.NewSource(1))}

	req := GetPrefectureAtRequestObject{Params: GetPrefectureAtParams{Lat: 35.735, Lng: 139.65}}
	resp, err := h.GetPrefectureAt(context.Background(), req)
	if err != nil {
		t.Fatalf("err = %v, want nil（型付き500で返すべき）", err)
	}
	if _, ok := resp.(GetPrefectureAt500JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want GetPrefectureAt500JSONResponse", resp)
	}
}

func TestGetPrefectureAt_repoエラーは型付き500を返す(t *testing.T) {
	h := NewHandler(&fakeRepo{err: errors.New("db 接続失敗")})

	req := GetPrefectureAtRequestObject{Params: GetPrefectureAtParams{Lat: 35.735, Lng: 139.65}}
	resp, err := h.GetPrefectureAt(context.Background(), req)
	if err != nil {
		t.Fatalf("err = %v, want nil（型付き500で返すべき）", err)
	}
	if _, ok := resp.(GetPrefectureAt500JSONResponse); !ok {
		t.Fatalf("レスポンス型 = %T, want GetPrefectureAt500JSONResponse", resp)
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
