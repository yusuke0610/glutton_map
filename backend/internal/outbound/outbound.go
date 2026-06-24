// Package outbound は公式 URL への計測付き送客（アウトバウンドクリック）を扱う。
//
// マップから如月公式への外部リンクは自前 backend 経由(GET /out)にしてクリックを記録し、
// 「マップが公式に N 件送客した」という計測可能な実績を残す。リダイレクト先は必ず
// サーバ側のホワイトリストに対応付け、クエリの任意 URL へは飛ばさない（オープンリダイレクト防止）。
//
// JSON ではない(302 リダイレクト)ため oapi-codegen の strict-server には乗せず、素 Gin ルートで登録する。
package outbound

import (
	"context"
	"time"
)

// Destinations は送客先のホワイトリスト。キー(to パラメータ) → 既知の公式 URL。
// ここに無いキーは 404 にする。任意 URL へのリダイレクトは許さない。
//
// TODO(運用): 実際のくいしんぼ如月公式 URL に差し替える（現状はプレースホルダ）。
// 次のアクションに近い URL を優先する（注文・メニュー・公式 SNS フォロー > トップページ）。
var Destinations = map[string]string{
	"official_menu": "http://www.nanban-tabetai.jp/",
	"official_sns":  "https://www.kisaragi-bento.example.com/sns",
}

// Resolve はホワイトリストから送客先 URL を引く。未知のキーは ok=false。
func Resolve(to string) (url string, ok bool) {
	url, ok = Destinations[to]
	return
}

// Click は1件のアウトバウンドクリック。後から復元できない流入元・時刻を記録する。
type Click struct {
	// Destination はホワイトリストのキー（official_menu 等）。生 URL ではなくキーを残す。
	Destination string
	UTMSource   string
	UTMMedium   string
	UTMCampaign string
	// CreatedAt は UTC で記録する。
	CreatedAt time.Time
}

// ClickRepository はクリック記録の永続化 seam。実装は DB を知る唯一のファイル
// (internal/pin/repository.go) に置き、DB 隔離原則を守る。
type ClickRepository interface {
	RecordClick(ctx context.Context, c Click) error
}
