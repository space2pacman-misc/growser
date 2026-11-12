package ui

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	"growser/colorutil"
	"growser/layout"
)

var face = basicfont.Face7x13

var (
	toolbarBg   = color.RGBA{235, 235, 240, 255}
	toolbarLine = color.RGBA{200, 200, 210, 255}
	inputBg     = color.RGBA{255, 255, 255, 255}
	inputBorder = color.RGBA{160, 160, 170, 255}
	textDark    = color.RGBA{30, 30, 30, 255}
	textGray    = color.RGBA{120, 120, 120, 255}
	linkBlue    = color.RGBA{0, 0, 200, 255}
)

func (a *App) Draw(screen *ebiten.Image) {
	// фон страницы
	bg := color.RGBA{255, 255, 255, 255}
	if a.Root.BgColor != "" {
		if c, ok := colorutil.Parse(a.Root.BgColor); ok {
			bg = c
		}
	}
	screen.Fill(bg)

	// страница (со сдвигом вниз на ToolbarH)
	a.drawBox(screen, a.Root, a.ScrollY)

	// панель сверху (рисуем поверх страницы)
	a.drawToolbar(screen)

	// статус-бар внизу
	if a.hoverLink != "" || a.Status != "" {
		msg := a.hoverLink
		if msg == "" {
			msg = a.Status
		}
		vector.DrawFilledRect(screen, 0, WindowHeight-24, WindowWidth, 24,
			color.RGBA{240, 240, 240, 255}, false)
		text.Draw(screen, msg, face, 8, WindowHeight-6, textDark)
	}
}

// drawToolbar — верхняя панель с кнопками и полем ввода.
func (a *App) drawToolbar(screen *ebiten.Image) {
	vector.DrawFilledRect(screen, 0, 0, WindowWidth, ToolbarH, toolbarBg, false)
	vector.StrokeLine(screen, 0, ToolbarH, WindowWidth, ToolbarH, 1, toolbarLine, false)

	// кнопки ← →
	drawArrow(screen, 8, ToolbarH/2, "left", a.HistoryI > 0)
	drawArrow(screen, ToolbarH+8, ToolbarH/2, "right", a.HistoryI < len(a.History)-1)

	// поле ввода
	fieldX := ToolbarH*2 + 4
	fieldY := 6
	fieldW := WindowWidth - fieldX - 8
	fieldH := ToolbarH - 12
	vector.DrawFilledRect(screen, float32(fieldX), float32(fieldY), float32(fieldW), float32(fieldH), inputBg, false)
	border := inputBorder
	if a.editing {
		border = color.RGBA{80, 120, 220, 255}
	}
	vector.StrokeRect(screen, float32(fieldX), float32(fieldY), float32(fieldW), float32(fieldH), 1, border, false)

	// текст
	txt := a.URL
	if a.editing {
		txt = a.editText
	}

	// скроллим текст по горизонтали, если не влезает
	const charW = 7
	maxChars := (fieldW - 12) / charW
	runes := []rune(txt)
	start := 0
	if a.editing && a.caret > maxChars {
		start = a.caret - maxChars
	}
	if start > len(runes) {
		start = len(runes)
	}
	visible := runes[start:]
	if len(visible) > maxChars {
		visible = visible[:maxChars]
	}

	tx := fieldX + 6
	ty := fieldY + (fieldH+13)/2 - 1
	text.Draw(screen, string(visible), face, tx, ty, textDark)

	// курсор
	if a.editing {
		caretVisible := (a.caret - start) * charW
		if caretVisible < 0 {
			caretVisible = 0
		}
		cx := tx + caretVisible
		vector.StrokeLine(screen, float32(cx), float32(fieldY+4),
			float32(cx), float32(fieldY+fieldH-4), 1, textDark, false)
	}
}

// drawArrow — треугольник-стрелка. Кнопка в квадрате ToolbarH.
func drawArrow(screen *ebiten.Image, x, cy int, dir string, enabled bool) {
	c := textDark
	if !enabled {
		c = textGray
	}
	s := float32(6)
	cx := float32(x + 4)
	fcy := float32(cy)
	if dir == "left" {
		vector.StrokeLine(screen, cx-s, fcy, cx+s, fcy-s, 2, c, true)
		vector.StrokeLine(screen, cx-s, fcy, cx+s, fcy+s, 2, c, true)
	} else {
		vector.StrokeLine(screen, cx+s, fcy, cx-s, fcy-s, 2, c, true)
		vector.StrokeLine(screen, cx+s, fcy, cx-s, fcy+s, 2, c, true)
	}
}

func (a *App) drawBox(screen *ebiten.Image, b *layout.Box, scrollY float64) {
	// всё рисуем со сдвигом +ToolbarH
	offsetY := float32(ToolbarH)

	// подсветка hover
	if b.Href != "" && b.Href == a.hoverLink {
		vector.DrawFilledRect(screen,
			float32(b.X-b.PaddingLeft),
			float32(b.Y-b.PaddingTop-scrollY)+offsetY,
			float32(b.Width+b.PaddingLeft+b.PaddingRight),
			float32(b.Height+b.PaddingTop+b.PaddingBottom),
			color.RGBA{255, 255, 200, 255}, false)
	}

	if b.BgColor != "" {
		if c, ok := colorutil.Parse(b.BgColor); ok {
			x := float32(b.X - b.PaddingLeft)
			y := float32(b.Y-b.PaddingTop-scrollY) + offsetY
			w := float32(b.Width + b.PaddingLeft + b.PaddingRight)
			h := float32(b.Height + b.PaddingTop + b.PaddingBottom)
			if h > 0 && w > 0 {
				vector.DrawFilledRect(screen, x, y, w, h, c, false)
			}
		}
	}

	if b.Text != "" {
		colorOverride := color.RGBA{}
		if b.Href != "" {
			colorOverride = linkBlue
		}
		a.drawTextBlock(screen, b, scrollY, colorOverride)
	}

	for _, c := range b.Children {
		a.drawBox(screen, c, scrollY)
	}
}

func (a *App) drawTextBlock(screen *ebiten.Image, b *layout.Box, scrollY float64, override color.RGBA) {
	const charW = 7
	maxChars := int(b.Width) / charW
	if maxChars < 1 {
		maxChars = 1
	}

	textColor := textDark
	if b.Color != "" {
		if c, ok := colorutil.Parse(b.Color); ok {
			textColor = c
		}
	}
	if override != (color.RGBA{}) {
		textColor = override
	}

	lineY := int(b.Y-scrollY) + 13 + ToolbarH
	var line []rune

	flush := func() {
		if len(line) == 0 {
			return
		}
		lineW := len(line) * charW
		x := int(b.X)
		switch b.TextAlign {
		case "center":
			x = int(b.X) + (int(b.Width)-lineW)/2
		case "right":
			x = int(b.X) + int(b.Width) - lineW
		}
		text.Draw(screen, string(line), face, x, lineY, textColor)
		if b.Href != "" {
			vector.StrokeLine(screen,
				float32(x), float32(lineY+2),
				float32(x+lineW), float32(lineY+2),
				1, textColor, false)
		}
		lineY += int(b.LineHeight)
		if lineY < 13 {
			lineY = 13
		}
		line = line[:0]
	}

	for _, word := range splitWords(b.Text) {
		r := []rune(word)
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

func splitWords(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' || r == '\r' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}