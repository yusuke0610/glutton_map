// Package share は X 共有のための SSR ルートを提供する。
//
// X(Twitter) のクローラは JS を実行しないため、クライアントレンダリングの SPA を
// そのまま共有するとカードが空になる。そこで共有リンクはこの /share（Go が meta 入りの
// HTML をサーバーサイドで返す素 Gin ルート）を指し、クローラには OGP/Twitter Card を、
// 人間にはフロント(SPA)の「刺せる状態のマップ」へのリダイレクトを返す。
//
// これらは JSON ではないため oapi-codegen の strict-server には乗せない（契約外の素ルート）。
package share

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/kisaragi-ai-map/backend/internal/pin"
)

// ogpImage は同梱の静的 OGP 画像（1200x630）。動的（県名・ピン数焼き込み）は別 PR。
//
//go:embed assets/ogp.png
var ogpImage []byte

// Config は SSR 共有ルートの依存。
type Config struct {
	// PublicBaseURL は backend 自身の公開 URL（https）。og:image / og:url を絶対化するのに使う。
	// X bot は相対パス・localhost・ログイン必須ページを取得できないため絶対 https にする。
	PublicBaseURL string
	// FrontendBaseURL は SPA の公開 URL。人間向け redirect 先。
	FrontendBaseURL string
}

// UTM は共有リンクに付く流入元の計測値。redirect 先へ引き継ぎ、投稿時に記録させる。
type UTM struct {
	Source   string
	Medium   string
	Campaign string
}

// Handler は /share と /static/ogp.png を提供する。
type Handler struct {
	cfg  Config
	tmpl *template.Template
}

func NewHandler(cfg Config) *Handler {
	return &Handler{cfg: cfg, tmpl: pageTmpl}
}

// Register は素 Gin ルートを登録する（strict-server とは別系統）。
func (h *Handler) Register(r gin.IRouter) {
	r.GET("/share", h.share)
	r.GET("/static/ogp.png", h.ogp)
}

func (h *Handler) share(c *gin.Context) {
	pref := c.Query("pref")
	utm := UTM{
		Source:   c.Query("utm_source"),
		Medium:   c.Query("utm_medium"),
		Campaign: c.Query("utm_campaign"),
	}
	html, err := h.renderHTML(pref, utm)
	if err != nil {
		c.String(http.StatusInternalServerError, "共有ページの生成に失敗しました")
		return
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *Handler) ogp(c *gin.Context) {
	// X はカードをキャッシュするため、将来動的化したら共有 URL に ?v= を付けて再スクレイプさせる。
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", ogpImage)
}

// renderHTML は OGP/Twitter Card 入りの HTML を組み立てる純粋寄りのメソッド。
// pref はホワイトリスト（47都道府県）外なら無視して汎用文面にフォールバックする
// （反射型インジェクション・表記ゆれ対策）。
func (h *Handler) renderHTML(pref string, utm UTM) (string, error) {
	valid := isPrefecture(pref)

	title := "くいしんぼ如月ファンマップ"
	desc := "全国のくいしんぼ如月ファンが集まる地図。あなたもピンを刺して仲間を増やそう🍱"
	if valid {
		title = pref + "のくいしんぼ如月ファンマップ"
		desc = pref + "のファンが集まる地図。あなたもピンを刺して仲間を増やそう🍱"
	}

	prefForLink := ""
	if valid {
		prefForLink = pref
	}

	data := pageData{
		Title:        title,
		Description:  desc,
		ImageURL:     h.imageURL(),
		CanonicalURL: h.canonicalURL(prefForLink),
		RedirectURL:  h.redirectURL(prefForLink, utm),
	}

	var buf bytes.Buffer
	if err := h.tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (h *Handler) imageURL() string {
	return strings.TrimRight(h.cfg.PublicBaseURL, "/") + "/static/ogp.png"
}

func (h *Handler) canonicalURL(pref string) string {
	u := strings.TrimRight(h.cfg.PublicBaseURL, "/") + "/share"
	if pref != "" {
		u += "?" + url.Values{"pref": {pref}}.Encode()
	}
	return u
}

// redirectURL は人間向けの着地先（フロント）を作る。着地後すぐ刺せるよう post=1 を付け、
// 計測のため utm をそのまま引き継ぐ。
func (h *Handler) redirectURL(pref string, utm UTM) string {
	q := url.Values{}
	q.Set("post", "1")
	if pref != "" {
		q.Set("pref", pref)
	}
	if utm.Source != "" {
		q.Set("utm_source", utm.Source)
	}
	if utm.Medium != "" {
		q.Set("utm_medium", utm.Medium)
	}
	if utm.Campaign != "" {
		q.Set("utm_campaign", utm.Campaign)
	}
	return strings.TrimRight(h.cfg.FrontendBaseURL, "/") + "/?" + q.Encode()
}

type pageData struct {
	Title        string
	Description  string
	ImageURL     string
	CanonicalURL string
	RedirectURL  string
}

// pageTmpl は html/template でコンテキストに応じたエスケープを行う（XSS 防止）。
var pageTmpl = template.Must(template.New("share").Parse(`<!doctype html>
<html lang="ja">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<meta name="twitter:card" content="summary_large_image">
<meta name="twitter:title" content="{{.Title}}">
<meta name="twitter:description" content="{{.Description}}">
<meta name="twitter:image" content="{{.ImageURL}}">
<meta property="og:type" content="website">
<meta property="og:title" content="{{.Title}}">
<meta property="og:description" content="{{.Description}}">
<meta property="og:image" content="{{.ImageURL}}">
<meta property="og:url" content="{{.CanonicalURL}}">
<meta http-equiv="refresh" content="0;url={{.RedirectURL}}">
<script>location.replace({{.RedirectURL}})</script>
</head>
<body>
<p>地図へ移動します… <a href="{{.RedirectURL}}">開かない場合はこちら</a></p>
</body>
</html>
`))

// prefectureSet は pin の重心マップ（47都道府県）から作るホワイトリスト。
var prefectureSet = func() map[string]bool {
	m := map[string]bool{}
	pin.EachPrefecture(func(p pin.Prefecture) { m[string(p)] = true })
	return m
}()

func isPrefecture(name string) bool { return prefectureSet[name] }
