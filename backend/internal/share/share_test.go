package share

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func testHandler() *Handler {
	return NewHandler(Config{
		PublicBaseURL:   "https://api.example.com",
		FrontendBaseURL: "https://map.example.com",
	})
}

func TestRenderHTML_summary_large_imageと絶対httpsのog_image(t *testing.T) {
	h := testHandler()
	html, err := h.renderHTML("高知県", UTM{Source: "twitter", Medium: "social", Campaign: "fan_share"})
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}

	// X クローラ向けにツイート幅いっぱいのカードを明示する。
	if !strings.Contains(html, `name="twitter:card" content="summary_large_image"`) {
		t.Errorf("summary_large_image が無い:\n%s", html)
	}
	// og:image / twitter:image は絶対 https・公開 URL であること。
	wantImg := "https://api.example.com/static/ogp.png"
	if !strings.Contains(html, `property="og:image" content="`+wantImg+`"`) {
		t.Errorf("og:image = 絶対https を期待: %s が無い", wantImg)
	}
	if !strings.Contains(html, `name="twitter:image" content="`+wantImg+`"`) {
		t.Errorf("twitter:image = 絶対https を期待: %s が無い", wantImg)
	}
	// OGP 基本タグも併記する（X は twitter 専用タグが無ければ OGP にフォールバック）。
	for _, want := range []string{`property="og:title"`, `property="og:description"`, `property="og:url"`, `property="og:type"`} {
		if !strings.Contains(html, want) {
			t.Errorf("%s が無い", want)
		}
	}
}

func TestRenderHTML_有効な県名は文面に入る(t *testing.T) {
	h := testHandler()
	html, err := h.renderHTML("高知県", UTM{})
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}
	if !strings.Contains(html, "高知県") {
		t.Errorf("県名が文面に入っていない:\n%s", html)
	}
}

func TestRenderHTML_不正な県名は無視して汎用文面_インジェクション防止(t *testing.T) {
	h := testHandler()
	html, err := h.renderHTML(`"><script>alert(1)</script>`, UTM{})
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}
	// ホワイトリスト外はそのまま反映しない（汎用文面へフォールバック）。
	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Errorf("未エスケープのスクリプトが混入:\n%s", html)
	}
}

func TestRenderHTML_リダイレクト先にutmとpostを引き継ぐ(t *testing.T) {
	h := testHandler()
	html, err := h.renderHTML("高知県", UTM{Source: "twitter", Medium: "social", Campaign: "fan_share"})
	if err != nil {
		t.Fatalf("renderHTML: %v", err)
	}
	// 人間向け redirect 先はフロント。着地後すぐ刺せるよう post=1、計測のため utm を引き継ぐ。
	for _, want := range []string{"map.example.com", "post=1", "utm_source=twitter", "utm_medium=social", "utm_campaign=fan_share"} {
		if !strings.Contains(html, want) {
			t.Errorf("redirect 先に %q が無い:\n%s", want, html)
		}
	}
}

func TestHTTP_Share_200でHTMLを返す(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	testHandler().Register(r)

	req := httptest.NewRequest(http.MethodGet, "/share?pref=高知県&utm_source=twitter", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (body=%s)", w.Code, w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	if !strings.Contains(w.Body.String(), "summary_large_image") {
		t.Error("body に OGP カード指定が無い")
	}
}

func TestHTTP_OGP画像_image_pngで返す(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	testHandler().Register(r)

	req := httptest.NewRequest(http.MethodGet, "/static/ogp.png", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
	if w.Body.Len() == 0 {
		t.Error("画像ボディが空")
	}
	// PNG シグネチャで始まること。
	if got := w.Body.Bytes(); len(got) < 8 || string(got[1:4]) != "PNG" {
		t.Errorf("PNG シグネチャでない: %v", got[:min(8, len(got))])
	}
}
