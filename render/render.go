package render

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"growser/layout"
)

var Debug bool

var (
	bgColor   = color.RGBA{255, 255, 255, 255}
	textColor = color.RGBA{20, 20, 20, 255}
	boxColor  = color.RGBA{230, 230, 240, 255}
)

func Render(root *layout.Box, path string) error {
	pageH := pageHeight(root) + 20
	if pageH < 600 {
		pageH = 600
	}
	w := int(layout.ViewportWidth)
	h := int(pageH)
	if h < int(layout.ViewportHeight) {
		h = int(layout.ViewportHeight)
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Фон: если у корня задан background — он заливает всю страницу
	fillColor := bgColor
	if root.BgColor != "" {
		if c, ok := parseColor(root.BgColor); ok {
			fillColor = c
		}
	}
	draw.Draw(img, img.Bounds(), &image.Uniform{fillColor}, image.Point{}, draw.Src)

	drawBox(img, root)

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func pageHeight(b *layout.Box) float64 {
	maxY := b.Y + b.Height + b.PaddingBottom + b.MarginBottom
	for _, c := range b.Children {
		if h := pageHeight(c); h > maxY {
			maxY = h
		}
	}
	return maxY
}

func drawBox(img *image.RGBA, b *layout.Box) {
	// фон
	if b.BgColor != "" {
		if c, ok := parseColor(b.BgColor); ok {
			x1 := int(b.X - b.PaddingLeft)
			y1 := int(b.Y - b.PaddingTop)
			x2 := int(b.X + b.Width + b.PaddingRight)
			y2 := int(b.Y + b.Height + b.PaddingBottom)
			fillRect(img, x1, y1, x2, y2, c)
		}
	}

	// отладочная рамка — можешь закомментировать
	if Debug && b.Width > 0 && b.Height > 0 {
		strokeRect(img, int(b.X), int(b.Y),
			int(b.X+b.Width), int(b.Y+b.Height), boxColor)
	}

	if b.Text != "" {
		drawText(img, b.Text, b)
	}

	for _, c := range b.Children {
		drawBox(img, c)
	}
}

func drawText(img *image.RGBA, text string, b *layout.Box) {
	face := basicfont.Face7x13
	charW := 7.0
	maxChars := int(b.Width / charW)
	if maxChars < 1 {
		maxChars = 1
	}

	lineY := int(b.Y) + 13
	var line []rune

	flush := func() {
		if len(line) == 0 {
			return
		}
		lineW := len(line) * 7
		x := int(b.X)
		switch b.TextAlign {
		case "center":
			x = int(b.X) + (int(b.Width)-lineW)/2
		case "right":
			x = int(b.X) + int(b.Width) - lineW
		}
		if x < 0 {
			x = 0
		}
		d := &font.Drawer{
			Dst:  img,
			Src:  image.NewUniform(textColor),
			Face: face,
			Dot:  fixed.P(x, lineY),
		}
		d.DrawString(string(line))
		lineY += int(b.LineHeight)
		if lineY < 13 {
			lineY = 13
		}
		line = line[:0]
	}

	for _, w := range strings.Fields(text) {
		r := []rune(w)
		if len(line) == 0 {
			line = append(line, r...)
		} else if len(line)+1+len(r) <= maxChars {
			line = append(line, ' ')
			line = append(line, r...)
		} else {
			flush()
			line = append(line, r...)
		}
	}
	flush()
}

func fillRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	if x1 < 0 {
		x1 = 0
	}
	if y1 < 0 {
		y1 = 0
	}
	if x2 > img.Bounds().Dx() {
		x2 = img.Bounds().Dx()
	}
	if y2 > img.Bounds().Dy() {
		y2 = img.Bounds().Dy()
	}
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			img.Set(x, y, c)
		}
	}
}

func strokeRect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for x := x1; x <= x2; x++ {
		set(img, x, y1, c)
		set(img, x, y2, c)
	}
	for y := y1; y <= y2; y++ {
		set(img, x1, y, c)
		set(img, x2, y, c)
	}
}

func set(img *image.RGBA, x, y int, c color.Color) {
	if image.Pt(x, y).In(img.Bounds()) {
		img.Set(x, y, c)
	}
}

// parseColor поддерживает #rgb, #rrggbb, несколько именованных.
func parseColor(s string) (color.RGBA, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "white", "#fff", "#ffffff":
		return color.RGBA{255, 255, 255, 255}, true
	case "black", "#000", "#000000":
		return color.RGBA{0, 0, 0, 255}, true
	case "red":
		return color.RGBA{255, 0, 0, 255}, true
	case "blue":
		return color.RGBA{0, 0, 255, 255}, true
	case "green":
		return color.RGBA{0, 128, 0, 255}, true
	case "gray", "grey":
		return color.RGBA{128, 128, 128, 255}, true
	}
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) == 3 {
			r, _ := strconv.ParseUint(string(h[0])+string(h[0]), 16, 8)
			g, _ := strconv.ParseUint(string(h[1])+string(h[1]), 16, 8)
			b, _ := strconv.ParseUint(string(h[2])+string(h[2]), 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true
		}
		if len(h) == 6 {
			r, _ := strconv.ParseUint(h[0:2], 16, 8)
			g, _ := strconv.ParseUint(h[2:4], 16, 8)
			b, _ := strconv.ParseUint(h[4:6], 16, 8)
			return color.RGBA{uint8(r), uint8(g), uint8(b), 255}, true
		}
	}
	return color.RGBA{}, false
}