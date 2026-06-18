package db

import (
	"context"
	"math/rand"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// prefectureCentroids: {lat, lng}。openapi.yaml の Prefecture enum とキーを一致させる。
var prefectureCentroids = map[string][2]float64{
	"北海道": {43.06, 141.35}, "青森県": {40.82, 140.74}, "岩手県": {39.70, 141.15},
	"宮城県": {38.27, 140.87}, "秋田県": {39.72, 140.10}, "山形県": {38.24, 140.36},
	"福島県": {37.75, 140.47}, "茨城県": {36.34, 140.45}, "栃木県": {36.57, 139.88},
	"群馬県": {36.39, 139.06}, "埼玉県": {35.86, 139.65}, "千葉県": {35.61, 140.12},
	"東京都": {35.69, 139.69}, "神奈川県": {35.45, 139.64}, "新潟県": {37.90, 139.02},
	"富山県": {36.70, 137.21}, "石川県": {36.59, 136.63}, "福井県": {36.07, 136.22},
	"山梨県": {35.66, 138.57}, "長野県": {36.65, 138.18}, "岐阜県": {35.39, 136.72},
	"静岡県": {34.98, 138.38}, "愛知県": {35.18, 136.91}, "三重県": {34.73, 136.51},
	"滋賀県": {35.00, 135.87}, "京都府": {35.02, 135.76}, "大阪府": {34.69, 135.50},
	"兵庫県": {34.69, 134.04}, "愛媛県": {33.84, 132.77}, "高知県": {33.56, 133.53},
	"福岡県": {33.61, 130.42}, "佐賀県": {33.25, 130.30}, "長崎県": {32.74, 129.87},
	"熊本県": {32.79, 130.74}, "大分県": {33.24, 131.61}, "宮崎県": {31.91, 131.42},
	"鹿児島県": {31.56, 130.56}, "沖縄県": {26.21, 127.68},
}

// seedWeights: 件数（＝ヒートの濃さ）。地元・県外難民クラスタを表現。未指定は 1。
var seedWeights = map[string]int{
	"高知県": 14, "東京都": 10, "大阪府": 7, "神奈川県": 5,
	"愛知県": 3, "兵庫県": 3, "福岡県": 2, "京都府": 2, "千葉県": 2, "埼玉県": 2,
}

// Seed は DB が空のとき、重心+ゆらぎで pins を投入する。
func Seed(ctx context.Context, repo pin.PinRepository) error {
	n, err := repo.Count(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	for pref, c := range prefectureCentroids {
		count := seedWeights[pref]
		if count == 0 {
			count = 1
		}
		for i := 0; i < count; i++ {
			p := pin.Pin{
				Prefecture: pin.Prefecture(pref),
				Lat:        c[0] + (rand.Float64()*0.3 - 0.15),
				Lng:        c[1] + (rand.Float64()*0.3 - 0.15),
			}
			if err := repo.Insert(ctx, p); err != nil {
				return err
			}
		}
	}
	return nil
}
