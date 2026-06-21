package db

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// seedWeights: 件数（＝ヒートの濃さ）。地元・県外難民クラスタを表現。未指定は 1。
var seedWeights = map[pin.Prefecture]int{
	"高知県": 14, "東京都": 10, "大阪府": 7, "神奈川県": 5,
	"愛知県": 3, "兵庫県": 3, "福岡県": 2, "京都府": 2, "千葉県": 2, "埼玉県": 2,
}

// Seed は DB が空のとき、重心+ゆらぎで pins を投入する。
// 座標生成は pin.CoordinateFor に一本化している（投稿ピンと同じロジック）。
func Seed(ctx context.Context, repo pin.PinRepository) error {
	n, err := repo.Count(ctx)
	if err != nil {
		return fmt.Errorf("seed 件数確認: %w", err)
	}
	if n > 0 {
		return nil
	}
	r := rand.New(rand.NewSource(rand.Int63()))
	var insertErr error
	pin.EachPrefecture(func(pref pin.Prefecture) {
		if insertErr != nil {
			return
		}
		count := seedWeights[pref]
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			lat, lng, ok := pin.CoordinateFor(pref, r)
			if !ok {
				continue
			}
			if err := repo.Insert(ctx, pin.Pin{Prefecture: pref, Lat: lat, Lng: lng}); err != nil {
				insertErr = fmt.Errorf("seed ピン投入: %w", err)
				return
			}
		}
	})
	return insertErr
}
