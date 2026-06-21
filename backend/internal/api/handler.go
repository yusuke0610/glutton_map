package api

import (
	"context"
	"log/slog"
	"math/rand"
	"sync"
	"time"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// Handler は生成された StrictServerInterface の実装。
type Handler struct {
	repo pin.PinRepository

	// rng は投稿ピンの座標生成に使う。*rand.Rand は並行安全でないため mu で保護する。
	mu  sync.Mutex
	rng *rand.Rand
}

func NewHandler(repo pin.PinRepository) *Handler {
	return &Handler{
		repo: repo,
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

// PostApiPins はファン投稿を1件受け取り、検証→座標生成→保存して 201 で返す。
// 座標はサーバが都道府県の重心+ゆらぎで生成し、クライアントの lat/lng は受け取らない。
func (h *Handler) PostApiPins(ctx context.Context, request PostApiPinsRequestObject) (PostApiPinsResponseObject, error) {
	if request.Body == nil {
		return PostApiPins400JSONResponse{Message: "リクエストボディがありません"}, nil
	}
	body := request.Body

	comment := ""
	if body.Comment != nil {
		comment = *body.Comment
	}

	if err := pin.ValidateCreate(body.Nickname, string(body.Prefecture), body.City, comment); err != nil {
		return PostApiPins400JSONResponse{Message: err.Error()}, nil
	}

	h.mu.Lock()
	lat, lng, ok := pin.CoordinateFor(pin.Prefecture(body.Prefecture), h.rng)
	h.mu.Unlock()
	if !ok {
		// ValidateCreate で都道府県は検証済みなので通常ここには来ない（防御的）。
		return PostApiPins400JSONResponse{Message: "都道府県が不正です"}, nil
	}

	p := pin.Pin{
		Prefecture: pin.Prefecture(body.Prefecture),
		Lat:        lat,
		Lng:        lng,
		Nickname:   body.Nickname,
		City:       body.City,
		Comment:    comment,
	}
	if err := h.repo.Insert(ctx, p); err != nil {
		slog.Error("ピン投稿の保存に失敗", "error", err)
		return PostApiPins500JSONResponse{Message: "ピンの投稿に失敗しました"}, nil
	}

	return PostApiPins201JSONResponse(toAPIPin(p)), nil
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
