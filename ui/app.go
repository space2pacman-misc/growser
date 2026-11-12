package ui

import (
	"fmt"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"growser/css"
	"growser/dom"
	"growser/layout"
	"growser/net"
)

const (
	WindowWidth  = 800
	WindowHeight = 600
	ToolbarH     = 36 // высота панели сверху
)

type App struct {
	Root    *layout.Box
	ScrollY float64

	URL      string
	History  []string
	HistoryI int
	Status   string

	// Ввод URL
	editing  bool   // фокус в адресной строке
	editText string // текущее содержимое поля
	caret    int    // позиция курсора

	hoverLink string
}

func NewApp(url string) (*App, error) {
	a := &App{URL: url, HistoryI: -1}
	if err := a.navigate(url, true); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) navigate(url string, pushHistory bool) error {
	body, err := net.Fetch(url)
	if err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	root, err := dom.Parse(body)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}

	var cssText strings.Builder
	var collect func(n *dom.Node)
	collect = func(n *dom.Node) {
		if n.Tag == "style" {
			cssText.WriteString(n.TextContent())
			cssText.WriteString("\n")
		}
		for _, c := range n.Children {
			collect(c)
		}
	}
	collect(root)
	rules := css.ParseStylesheet(cssText.String())

	a.Root = layout.Layout(root, rules)
	a.URL = url
	a.ScrollY = 0
	a.editing = false

	if pushHistory {
		if a.HistoryI >= 0 && a.HistoryI < len(a.History)-1 {
			a.History = a.History[:a.HistoryI+1]
		}
		a.History = append(a.History, url)
		a.HistoryI = len(a.History) - 1
	}

	ebiten.SetWindowTitle("Growser — " + url)
	return nil
}

func (a *App) Update() error {
	// === Режим ввода URL ===
	if a.editing {
		a.updateEditing()
		return nil
	}

	// === Обычный режим ===

	// Ctrl+L или просто L — фокус в адресную строку
	if inpututil.IsKeyJustPressed(ebiten.KeyL) &&
		(ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyAlt)) {
		a.startEditing()
		return nil
	}

	// колесо
	_, wy := ebiten.Wheel()
	a.ScrollY -= wy * 30

	// клавиши
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		a.ScrollY += 20
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		a.ScrollY -= 20
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		a.ScrollY += WindowHeight - ToolbarH - 40
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		a.ScrollY -= WindowHeight - ToolbarH - 40
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		a.ScrollY = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		a.ScrollY = pageHeight(a.Root)
	}

	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
	if alt && inpututil.IsKeyJustPressed(ebiten.KeyLeft) && a.HistoryI > 0 {
		a.HistoryI--
		a.navigate(a.History[a.HistoryI], false)
	}
	if alt && inpututil.IsKeyJustPressed(ebiten.KeyRight) && a.HistoryI < len(a.History)-1 {
		a.HistoryI++
		a.navigate(a.History[a.HistoryI], false)
	}

	mx, my := ebiten.CursorPosition()

	// Клик в панели — проверим кнопки/поле
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if my < ToolbarH {
			a.handleToolbarClick(mx, my)
			return nil
		}
	}

	a.updateHover(float64(mx), float64(my-ToolbarH)+a.ScrollY)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && my >= ToolbarH {
		a.handleClick(float64(mx), float64(my-ToolbarH))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}

	pageH := pageHeight(a.Root)
	maxScroll := pageH - float64(WindowHeight-ToolbarH)
	if maxScroll < 0 {
		maxScroll = 0
	}
	if a.ScrollY < 0 {
		a.ScrollY = 0
	}
	if a.ScrollY > maxScroll {
		a.ScrollY = maxScroll
	}

	return nil
}

// startEditing — войти в режим ввода URL.
func (a *App) startEditing() {
	a.editing = true
	a.editText = a.URL
	a.caret = len(a.editText)
	ebiten.SetCursorShape(ebiten.CursorShapeText)
}

// updateEditing — обработка ввода URL.
func (a *App) updateEditing() {
	// Выход
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.editing = false
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		return
	}

	// Ввод текста (в т.ч. кириллица, вставка)
	a.editText = string(ebiten.AppendInputChars([]rune(a.editText)))

	// Enter — навигация
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		url := strings.TrimSpace(a.editText)
		if url != "" && !strings.Contains(url, "://") {
			url = "https://" + url
		}
		a.editing = false
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
		if err := a.navigate(url, true); err != nil {
			a.Status = "error: " + err.Error()
		}
		return
	}

	// Backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		if a.caret > 0 {
			r := []rune(a.editText)
			a.editText = string(append(r[:a.caret-1], r[a.caret:]...))
			a.caret--
		}
	}

	// Delete
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		r := []rune(a.editText)
		if a.caret < len(r) {
			a.editText = string(append(r[:a.caret], r[a.caret+1:]...))
		}
	}

	// ← →
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && a.caret > 0 {
		a.caret--
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && a.caret < len([]rune(a.editText)) {
		a.caret++
	}

	// Home/End
	if inpututil.IsKeyJustPressed(ebiten.KeyHome) {
		a.caret = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnd) {
		a.caret = len([]rune(a.editText))
	}

	// Ctrl+A — выделить всё (просто курсор в конец — упрощение)
	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyA) {
		a.caret = len([]rune(a.editText))
	}
}

func (a *App) handleToolbarClick(mx, my int) {
	// кнопки ← → на панели
	if mx < ToolbarH && my < ToolbarH {
		if a.HistoryI > 0 {
			a.HistoryI--
			a.navigate(a.History[a.HistoryI], false)
		}
		return
	}
	if mx >= ToolbarH && mx < ToolbarH*2 && my < ToolbarH {
		if a.HistoryI < len(a.History)-1 {
			a.HistoryI++
			a.navigate(a.History[a.HistoryI], false)
		}
		return
	}
	// остальное — поле ввода: фокус
	a.startEditing()
}

func (a *App) updateHover(x, y float64) {
	a.hoverLink = ""
	hit := HitTest(a.Root, x, y)
	if hit != nil && hit.Href != "" {
		a.hoverLink = hit.Href
	}
	if a.hoverLink != "" {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}

func (a *App) handleClick(x, y float64) {
	pageY := y + a.ScrollY
	hit := HitTest(a.Root, x, pageY)
	if hit == nil || hit.Href == "" {
		return
	}
	target := hit.Href
	if strings.HasPrefix(target, "/") || !strings.HasPrefix(target, "http") && !strings.HasPrefix(target, "mailto:") {
		if u, err := resolveRelative(a.URL, target); err == nil {
			target = u
		}
	}
	if err := a.navigate(target, true); err != nil {
		fmt.Printf("navigate error: %v\n", err)
		a.Status = "error: " + err.Error()
	}
}

func (a *App) Layout(outsideWidth, outsideHeight int) (int, int) {
	return WindowWidth, WindowHeight
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