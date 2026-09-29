//go:build ignore

// 生成应用图标 build/appicon.png:蓝底圆角方块 + 白色闪电。
// 仅用标准库,逐像素绘制(一次性脚本):go run scripts/genicon.go
package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

const size = 512

func main() {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	bg := color.RGBA{R: 0x3e, G: 0x8e, B: 0xff, A: 0xff} // --color-accent #3e8eff
	fg := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff} // 白色闪电
	radius := 96.0                                       // 圆角

	// 闪电多边形(归一化坐标)
	bolt := [][2]float64{
		{0.60, 0.06}, {0.24, 0.56}, {0.46, 0.56}, {0.38, 0.94},
		{0.76, 0.42}, {0.53, 0.42},
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !inRoundedRect(float64(x)+0.5, float64(y)+0.5, radius) {
				continue // 圆角外保持透明
			}
			c := bg
			if inPolygon(float64(x)+0.5, float64(y)+0.5, bolt) {
				c = fg
			}
			img.SetRGBA(x, y, c)
		}
	}

	f, err := os.Create("build/appicon.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func inRoundedRect(px, py, r float64) bool {
	s := float64(size)
	cx := math.Max(r, math.Min(px, s-r))
	cy := math.Max(r, math.Min(py, s-r))
	dx, dy := px-cx, py-cy
	return dx*dx+dy*dy <= r*r
}

func inPolygon(px, py float64, poly [][2]float64) bool {
	inside := false
	j := len(poly) - 1
	for i := 0; i < len(poly); i++ {
		xi, yi := poly[i][0]*size, poly[i][1]*size
		xj, yj := poly[j][0]*size, poly[j][1]*size
		if (yi > py) != (yj > py) && px < (xj-xi)*(py-yi)/(yj-yi)+xi {
			inside = !inside
		}
		j = i
	}
	return inside
}
