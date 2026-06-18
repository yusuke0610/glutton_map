package api

import (
	"context"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// Handler は生成された StrictServerInterface の実装。
type Handler struct {
	repo pin.PinRepository
}

func NewHandler(repo pin.PinRepository) *Handler {
	return &Handler{repo: repo}
}

// GetApiPins はピン一覧と集計（prefecture_count / total）を返す。
func (h *Handler) GetApiPins(ctx context.Context, _ GetApiPinsRequestObject) (GetApiPinsResponseObject, error) {
	pins, err := h.repo.GetPins(ctx)
	if err != nil {
		return nil, err
	}

	weight := 1
	out := make([]Pin, 0, len(pins))
	seen := map[pin.Prefecture]struct{}{}
	for _, p := range pins {
		w := weight
		out = append(out, Pin{
			Prefecture: Prefecture(p.Prefecture),
			Lat:        p.Lat,
			Lng:        p.Lng,
			Weight:     &w,
		})
		seen[p.Prefecture] = struct{}{}
	}

	return GetApiPins200JSONResponse{
		Pins:            out,
		PrefectureCount: len(seen),
		Total:           len(out),
	}, nil
}
