//go:build ignore

// generate_ogp.go は静的 OGP 画像（assets/ogp.png）を生成する一回限りのスクリプト。
// 標準ライブラリのみで描画するため go.mod に依存を増やさない（go mod tidy で消えない）。
// 文字（県名・ピン数）の焼き込みは動的 OGP（フェーズ2/別PR）で対応する。ここでは
// summary_large_image 比率(1200x630)のブランド画像を作り、文言は meta タグ側に持たせる。
//
// 再生成: cd backend/internal/share/assets && go run generate_ogp.go
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const (
	width  = 1200
	height = 630
)

var (
	cream    = color.RGBA{0xFA, 0xF6, 0xEC, 0xFF} // 背景（温かみのあるクリーム）
	brand    = color.RGBA{0xD9, 0x7B, 0x3A, 0xFF} // くいしんぼ如月のブランドオレンジ
	pinRed   = color.RGBA{0xE0, 0x20, 0x2A, 0xFF} // 地図ピンの赤
	dotFaint = color.RGBA{0xD9, 0x7B, 0x3A, 0x55} // 散らばるピン（薄いオレンジ）
)

func main() {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(img, cream)

	// 下部のブランドバンド。
	for y := height - 70; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, brand)
		}
	}

	// 全国に散らばるピンを表す薄いドット群（密度＝件数のメタファ）。
	dots := [][2]int{
		{220, 180}, {300, 150}, {360, 220}, {180, 280}, {420, 300},
		{500, 200}, {560, 320}, {640, 260}, {700, 360}, {760, 240},
		{840, 340}, {900, 280}, {980, 380}, {280, 380}, {620, 420},
	}
	for _, d := range dots {
		fillCircle(img, d[0], d[1], 10, dotFaint)
	}

	// 主役の地図ピン（円＋下向き三角）。中央やや左に大きく配置する。
	cx, cy, r := 360, 300, 90
	fillCircle(img, cx, cy, r, pinRed)
	fillCircle(img, cx, cy, 34, cream) // ピンの穴
	fillTriangle(img, cx-r, cy+10, cx+r, cy+10, cx, cy+r+120, pinRed)

	out, err := os.Create("ogp.png")
	if err != nil {
		panic(err)
	}
	defer out.Close()
	if err := png.Encode(out, img); err != nil {
		panic(err)
	}
}

func fill(img *image.RGBA, c color.RGBA) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

// fillCircle は (cx,cy) 中心・半径 r の塗りつぶし円を描く。アンチエイリアスはしない。
func fillCircle(img *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := float64(x-cx), float64(y-cy)
			if dx*dx+dy*dy <= float64(r*r) {
				blend(img, x, y, c)
			}
		}
	}
}

// fillTriangle は3頂点の塗りつぶし三角形を描く（重心座標による内外判定）。
func fillTriangle(img *image.RGBA, x1, y1, x2, y2, x3, y3 int, c color.RGBA) {
	minX := min3(x1, x2, x3)
	maxX := max3(x1, x2, x3)
	minY := min3(y1, y2, y3)
	maxY := max3(y1, y2, y3)
	area := float64((x2-x1)*(y3-y1) - (x3-x1)*(y2-y1))
	if area == 0 {
		return
	}
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			w1 := float64((x2-x)*(y3-y)-(x3-x)*(y2-y)) / area
			w2 := float64((x3-x)*(y1-y)-(x1-x)*(y3-y)) / area
			w3 := 1 - w1 - w2
			if w1 >= 0 && w2 >= 0 && w3 >= 0 {
				blend(img, x, y, c)
			}
		}
	}
}

// blend は単純なアルファ合成で1ピクセルを描く。
func blend(img *image.RGBA, x, y int, c color.RGBA) {
	if x < 0 || y < 0 || x >= width || y >= height {
		return
	}
	if c.A == 0xFF {
		img.SetRGBA(x, y, c)
		return
	}
	bg := img.RGBAAt(x, y)
	a := float64(c.A) / 255
	img.SetRGBA(x, y, color.RGBA{
		R: uint8(math.Round(float64(c.R)*a + float64(bg.R)*(1-a))),
		G: uint8(math.Round(float64(c.G)*a + float64(bg.G)*(1-a))),
		B: uint8(math.Round(float64(c.B)*a + float64(bg.B)*(1-a))),
		A: 0xFF,
	})
}

func min3(a, b, c int) int { return minI(a, minI(b, c)) }
func max3(a, b, c int) int { return maxI(a, maxI(b, c)) }
func minI(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}
