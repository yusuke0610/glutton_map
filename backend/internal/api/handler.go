package api

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/kisaragi-ai-map/backend/internal/geo"
	"github.com/kisaragi-ai-map/backend/internal/httpmw"
	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// Handler は生成された StrictServerInterface の実装。
type Handler struct {
	repo pin.PinRepository

	// muni は市区町村の境界データ。コード指定時に境界内の座標を生成する。
	// ロードに失敗したら nil になり、座標は都道府県重心へフォールバックする。
	muni *geo.Index

	// rng は投稿ピンの座標生成に使う。*rand.Rand は並行安全でないため mu で保護する。
	mu  sync.Mutex
	rng *rand.Rand
}

func NewHandler(repo pin.PinRepository) *Handler {
	idx, err := geo.Default()
	if err != nil {
		// 同梱データの解析失敗は致命ではない（都道府県重心へフォールバックできる）。観測のためログする。
		slog.Error("市区町村境界データのロードに失敗（都道府県重心にフォールバックします）", "error", err)
	}
	return &Handler{
		repo: repo,
		muni: idx,
		rng:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GetApiPins はピン一覧と集計（prefecture_count / total）を返す。
func (h *Handler) GetApiPins(ctx context.Context, _ GetApiPinsRequestObject) (GetApiPinsResponseObject, error) {
	pins, err := h.repo.GetPins(ctx)
	if err != nil {
		// 内部エラーは Go の error として上げず、契約上の型付き 500 で返す。
		// 観測性のため発生箇所で原因をログする。
		slog.Error("ピン取得に失敗", "error", err)
		return GetApiPins500JSONResponse{Message: "ピンの取得に失敗しました"}, nil
	}

	out := make([]Pin, 0, len(pins))
	for _, p := range pins {
		out = append(out, toAPIPin(p))
	}

	summary := pin.Summarize(pins)
	return GetApiPins200JSONResponse{
		Pins:            out,
		PrefectureCount: summary.PrefectureCount,
		Total:           summary.Total,
	}, nil
}

// GetPrefectureAt はクリック地点(lat/lng)の都道府県を行政区域ポリゴンで判定し、
// その都道府県のピン合計件数を返す。地点がどの都道府県にも属さない(海上など)場合は 404。
func (h *Handler) GetPrefectureAt(ctx context.Context, request GetPrefectureAtRequestObject) (GetPrefectureAtResponseObject, error) {
	if h.muni == nil {
		// 境界データのロードに失敗した縮退モードでは判定できない（起動時にログ済み）。
		return GetPrefectureAt500JSONResponse{Message: "都道府県の判定ができません"}, nil
	}

	prefecture, ok := h.muni.PrefectureAt(request.Params.Lat, request.Params.Lng)
	if !ok {
		return GetPrefectureAt404JSONResponse{Message: "この地点に該当する都道府県がありません"}, nil
	}

	pins, err := h.repo.GetPins(ctx)
	if err != nil {
		slog.Error("ピン取得に失敗", "error", err)
		return GetPrefectureAt500JSONResponse{Message: "ピンの取得に失敗しました"}, nil
	}

	count := pin.CountByPrefecture(pins, pin.Prefecture(prefecture))
	return GetPrefectureAt200JSONResponse{
		Prefecture: Prefecture(prefecture),
		Count:      count,
	}, nil
}

// PostApiPins はファン投稿を1件受け取り、検証→座標生成→保存して 201 で返す。
// 座標はサーバが都道府県の重心+ゆらぎで生成し、クライアントの lat/lng は受け取らない。
func (h *Handler) PostApiPins(ctx context.Context, request PostApiPinsRequestObject) (PostApiPinsResponseObject, error) {
	if request.Body == nil {
		return PostApiPins400JSONResponse{Message: "リクエストボディがありません"}, nil
	}
	body := request.Body

	comment := strOrEmpty(body.Comment)

	if err := pin.ValidateCreate(body.Nickname, string(body.Prefecture), body.City, comment); err != nil {
		return PostApiPins400JSONResponse{Message: err.Error()}, nil
	}

	// 任意の流入計測フィールドはサーバ側で長さを検証する（strict-server は maxLength を強制しない）。
	anonToken := strOrEmpty(body.AnonToken)
	utmSource := strOrEmpty(body.UtmSource)
	utmMedium := strOrEmpty(body.UtmMedium)
	utmCampaign := strOrEmpty(body.UtmCampaign)
	if err := pin.ValidateInflow(anonToken, utmSource, utmMedium, utmCampaign); err != nil {
		return PostApiPins400JSONResponse{Message: err.Error()}, nil
	}

	lat, lng, city, err := h.resolveMunicipality(body)
	if err != nil {
		return PostApiPins400JSONResponse{Message: err.Error()}, nil
	}

	p := pin.Pin{
		Prefecture: pin.Prefecture(body.Prefecture),
		Lat:        lat,
		Lng:        lng,
		Nickname:   body.Nickname,
		City:       city,
		Comment:    comment,
		// 投稿は拒否せずそのまま保存し、提出用集計でユニーク化する。
		// ip_hash はミドルウェアが context に載せた匿名識別子（生IPは持たない）。
		IPHash: httpmw.IPHashFrom(ctx),
		// 都道府県コードはコード先頭2桁から導出（座標は境界内に生成済みなので point-in-polygon 不要）。
		PrefectureCode: pin.PrefectureCodeFromMunicipality(body.MunicipalityCode),
		// 後から復元できない流入元・匿名トークンを投稿の瞬間に保存する（任意項目）。
		AnonToken:   anonToken,
		UTMSource:   utmSource,
		UTMMedium:   utmMedium,
		UTMCampaign: utmCampaign,
	}
	if err := h.repo.Insert(ctx, p); err != nil {
		slog.Error("ピン投稿の保存に失敗", "error", err)
		return PostApiPins500JSONResponse{Message: "ピンの投稿に失敗しました"}, nil
	}

	return PostApiPins201JSONResponse(toAPIPin(p)), nil
}

// resolveMunicipality は投稿の座標と保存用 city を決める。
// あいまい検索の候補から選ばれた municipality_code を必須とし、コードが実在し、かつ
// 選択した都道府県に属することを検証する。通れば境界内に座標を生成し、city を正規名称で返す。
// 検証に通らないときは err を返し、呼び出し側が 400 で拒否する。
//
// 例外: 同梱の市区町村データのロードに失敗した縮退モード（h.muni == nil、起動時にログ済み）では
// 検証不能なため、システムを止めないよう都道府県の重心+ゆらぎにフォールバックする。
func (h *Handler) resolveMunicipality(body *CreatePinRequest) (lat, lng float64, city string, err error) {
	// municipality_code は必須。候補未選択は縮退モードかどうかに関わらず拒否し、
	// 縮退モードで空コードが 201 で通って必須契約を破らないようにする。
	if body.MunicipalityCode == "" {
		return 0, 0, "", fmt.Errorf("市区町村は候補から選択してください")
	}

	if h.muni == nil {
		h.mu.Lock()
		la, lo, ok := pin.CoordinateFor(pin.Prefecture(body.Prefecture), h.rng)
		h.mu.Unlock()
		if !ok {
			// ValidateCreate で都道府県は検証済みなので通常ここには来ない（防御的）。
			return 0, 0, "", fmt.Errorf("都道府県が不正です")
		}
		return la, lo, body.City, nil
	}

	m, found := h.muni.Get(body.MunicipalityCode)
	if !found {
		return 0, 0, "", fmt.Errorf("市区町村が見つかりません")
	}
	if m.Prefecture != string(body.Prefecture) {
		return 0, 0, "", fmt.Errorf("市区町村が選択した都道府県に属していません")
	}

	h.mu.Lock()
	la, lo, ok := h.muni.SamplePoint(body.MunicipalityCode, h.rng)
	h.mu.Unlock()
	if !ok {
		// SamplePoint は代表点フォールバックで常に true を返す想定（防御的）。
		return 0, 0, "", fmt.Errorf("座標の生成に失敗しました")
	}
	// 表示用 city はコードの正規名称で上書きする（表記ゆれ対策）。
	return la, lo, m.Name, nil
}

// toAPIPin はドメイン Pin を API レスポンスの Pin に変換する。
// 最小スライスでは weight は常に 1。空文字のフィールドは nil にして payload を軽くする
// （seed 由来のピンは nickname/city/comment を持たない）。
func toAPIPin(p pin.Pin) Pin {
	weight := 1
	return Pin{
		Prefecture: Prefecture(p.Prefecture),
		Lat:        p.Lat,
		Lng:        p.Lng,
		Weight:     &weight,
		Nickname:   nilIfEmpty(p.Nickname),
		City:       nilIfEmpty(p.City),
		Comment:    nilIfEmpty(p.Comment),
	}
}

// nilIfEmpty は空文字なら nil を返す（JSON で omitempty を効かせるため）。
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// strOrEmpty は任意ポインタ文字列を非ポインタへ変換する（nil は空文字）。
func strOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
