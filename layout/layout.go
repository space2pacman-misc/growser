package layout

import (
	"strings"

	"growser/css"
	"growser/dom"
)

const (
	ViewportWidth   = 800.0
	ViewportHeight  = 600.0
	DefaultFontSize = 16.0
)

var blockTags = map[string]bool{
	"html": true, "body": true, "div": true, "p": true,
	"h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true,
	"ul": true, "ol": true, "li": true, "section": true, "article": true,
	"header": true, "footer": true, "main": true, "nav": true,
	"blockquote": true, "pre": true, "hr": true,
}

var hiddenTags = map[string]bool{
	"head": true, "script": true, "style": true, "meta": true,
	"link": true, "title": true,
}

// Box — прямоугольник на странице.
type Box struct {
	Node *dom.Node
	Text string

	// Позиция и размеры контента (без margin/padding).
	X, Y          float64
	Width, Height float64

	MarginTop, MarginRight, MarginBottom, MarginLeft     float64
	PaddingTop, PaddingRight, PaddingBottom, PaddingLeft float64

	FontSize   float64
	LineHeight float64
	TextAlign  string
	Color      string
	BgColor    string

	Children []*Box
	Href     string
}

// OuterWidth — ширина с margin.
func (b *Box) OuterWidth() float64 {
	return b.Width + b.PaddingLeft + b.PaddingRight + b.MarginLeft + b.MarginRight
}

// OuterHeight — высота с margin.
func (b *Box) OuterHeight() float64 {
	return b.Height + b.PaddingTop + b.PaddingBottom + b.MarginTop + b.MarginBottom
}

// Layout строит дерево Box из DOM, применяя CSS.
func Layout(root *dom.Node, rules []css.Rule) *Box {
	rootProps := resolveProps(root, rules, nil)

	box := &Box{Node: root}
	applyBoxProps(box, rootProps, ViewportWidth, DefaultFontSize)

	box.X = box.MarginLeft + box.PaddingLeft
	box.Y = box.MarginTop + box.PaddingTop

	contentWidth := ViewportWidth - box.MarginLeft - box.MarginRight -
		box.PaddingLeft - box.PaddingRight
	box.Width = contentWidth

	layoutChildren(box, rules, rootProps, contentWidth, "")

	return box
}

// resolveProps вычисляет CSS-свойства для узла с наследованием.
func resolveProps(node *dom.Node, rules []css.Rule, parentProps map[string]string) map[string]string {
	inherited := map[string]string{}
	if parentProps != nil {
		for _, k := range []string{"color", "font-size", "font-family", "line-height", "text-align"} {
			if v, ok := parentProps[k]; ok {
				inherited[k] = v
			}
		}
	}

	classes := classSet(node)
	inline := ""
	if node.Attrs != nil {
		inline = node.Attrs["style"]
	}
	computed := css.ComputeProperties(rules, node.Tag, classes, inline)

	out := make(map[string]string, len(inherited)+len(computed))
	for k, v := range inherited {
		out[k] = v
	}
	for k, v := range computed {
		out[k] = v
	}
	return out
}

func classSet(node *dom.Node) map[string]bool {
	set := make(map[string]bool)
	if node.Attrs == nil {
		return set
	}
	if cls, ok := node.Attrs["class"]; ok {
		for _, c := range strings.Fields(cls) {
			set[c] = true
		}
	}
	if id, ok := node.Attrs["id"]; ok {
		set["#"+id] = true
	}
	return set
}

// applyBoxProps заполняет поля Box из CSS-свойств.
func applyBoxProps(box *Box, props map[string]string, parentWidth, parentFontSize float64) {
	// font-size
	box.FontSize = parentFontSize
	if v, ok := props["font-size"]; ok {
		ctx := css.Context{FontSize: parentFontSize, ViewportW: ViewportWidth, ViewportH: ViewportHeight, ParentWidth: parentWidth}
		box.FontSize = css.Length(v, ctx)
	}
	if box.FontSize <= 0 {
		box.FontSize = DefaultFontSize
	}

	// line-height
	box.LineHeight = 1.4 * box.FontSize
	if v, ok := props["line-height"]; ok {
		v = strings.TrimSpace(v)
		if isNumeric(v) {
			f, _ := parseFloat(v)
			box.LineHeight = box.FontSize * f
		} else {
			ctx := css.Context{FontSize: box.FontSize, ViewportW: ViewportWidth, ViewportH: ViewportHeight, ParentWidth: parentWidth}
			box.LineHeight = css.Length(v, ctx)
		}
	}

	// text-align / color
	box.TextAlign = "left"
	if v, ok := props["text-align"]; ok {
		box.TextAlign = v
	}
	box.Color = "#222"
	if v, ok := props["color"]; ok {
		box.Color = v
	}
	if v, ok := props["background"]; ok {
		box.BgColor = extractColor(v)
	}
	if v, ok := props["background-color"]; ok {
		box.BgColor = strings.TrimSpace(v)
	}

	ctx := css.Context{FontSize: box.FontSize, ViewportW: ViewportWidth, ViewportH: ViewportHeight, ParentWidth: parentWidth}

	box.MarginTop, box.MarginRight, box.MarginBottom, box.MarginLeft =
		parseSides(props["margin"], ctx, 0, 0, 0, 0)
	box.PaddingTop, box.PaddingRight, box.PaddingBottom, box.PaddingLeft =
		parseSides(props["padding"], ctx, 0, 0, 0, 0)
}

func parseSides(val string, ctx css.Context, defT, defR, defB, defL float64) (t, r, b, l float64) {
	val = strings.TrimSpace(val)
	if val == "" {
		return defT, defR, defB, defL
	}
	parts := strings.Fields(val)
	lens := make([]float64, len(parts))
	for i, p := range parts {
		lens[i] = css.Length(p, ctx)
	}
	switch len(lens) {
	case 1:
		return lens[0], lens[0], lens[0], lens[0]
	case 2:
		return lens[0], lens[1], lens[0], lens[1]
	case 3:
		return lens[0], lens[1], lens[2], lens[1]
	default:
		return lens[0], lens[1], lens[2], lens[3]
	}
}

func extractColor(v string) string {
	if strings.HasPrefix(v, "light-dark(") {
		inner := strings.TrimSuffix(strings.TrimPrefix(v, "light-dark("), ")")
		parts := strings.SplitN(inner, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	return strings.TrimSpace(v)
}

func isNumeric(s string) bool {
	_, err := parseFloat(s)
	return err == nil
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmtSscan(s, &f)
	return f, err
}

// layoutChildren раскладывает детей блока столбиком (блоки) + инлайн-потоки.
func layoutChildren(box *Box, rules []css.Rule, parentProps map[string]string, availableWidth float64, parentHref string) {
	box.Width = availableWidth // ← ключевой фикс: ширина контента блока
	curY := box.Y + box.PaddingTop
	var inlineRun []*dom.Node
	var inlineProps []map[string]string

	flushInline := func() {
		if len(inlineRun) == 0 {
			return
		}

		// 1. Разбиваем inline-поток на runs — куски с одинаковым href.
		type run struct {
			text string
			href string
		}
		var runs []run
		var cur run
		cur.href = parentHref

		flushCur := func() {
			cur.text = strings.TrimSpace(cur.text)
			if cur.text != "" {
				runs = append(runs, cur)
			}
			cur = run{href: parentHref}
		}

		for _, n := range inlineRun {
			if n.Tag == "a" {
				flushCur()
				href, _ := n.Attrs["href"]
				cur.href = href
				cur.text = n.TextContent()
				flushCur()
			} else {
				cur.text += n.TextContent()
			}
		}
		flushCur()

		inlineRun = nil
		inlineProps = nil

		if len(runs) == 0 {
			return
		}

		// 2. Собираем слова с привязкой к href.
		type word struct {
			text string
			href string
		}
		var words []word
		for _, r := range runs {
			for _, w := range strings.Fields(r.text) {
				words = append(words, word{text: w, href: r.href})
			}
		}
		if len(words) == 0 {
			return
		}

		// 3. Раскладываем в строки и группируем подряд идущие слова одного href.
		innerW := availableWidth - box.PaddingLeft - box.PaddingRight
		charW := box.FontSize * 0.5
		if charW <= 0 {
			charW = 7
		}
		maxChars := int(innerW / charW)
		if maxChars < 1 {
			maxChars = 1
		}

		lineY := curY
		var line []word
		lineLen := 0

		flushLine := func() {
			if len(line) == 0 {
				return
			}
			x := box.X + box.PaddingLeft
			i := 0
			for i < len(line) {
				j := i
				for j < len(line) && line[j].href == line[i].href {
					j++
				}
				var sb strings.Builder
				for k := i; k < j; k++ {
					if k > i {
						sb.WriteByte(' ')
					}
					sb.WriteString(line[k].text)
				}
				fragText := sb.String()
				fragW := float64(len([]rune(fragText))) * charW

				box.Children = append(box.Children, &Box{
					Text:       fragText,
					X:          x,
					Y:          lineY,
					Width:      fragW,
					Height:     box.LineHeight + 4,
					FontSize:   box.FontSize,
					LineHeight: box.LineHeight,
					TextAlign:  "left",
					Color:      box.Color,
					Href:       line[i].href,
				})
				x += fragW + charW
				i = j
			}
			lineY += box.LineHeight
			line = nil
			lineLen = 0
		}

		for _, w := range words {
			wl := len([]rune(w.text))
			if lineLen > 0 && lineLen+1+wl > maxChars {
				flushLine()
			}
			line = append(line, w)
			if lineLen == 0 {
				lineLen = wl
			} else {
				lineLen += 1 + wl
			}
		}
		flushLine()

		curY = lineY
	}

	for _, child := range box.Node.Children {
		if hiddenTags[child.Tag] {
			continue
		}
		childProps := resolveProps(child, rules, parentProps)

		childHref := parentHref
		if child.Tag == "a" {
			if href, ok := child.Attrs["href"]; ok && href != "" {
				childHref = href
			}
		}

		if blockTags[child.Tag] || child.Tag == "#document" {
			flushInline()

			cb := &Box{Node: child, Href: childHref}
			applyBoxProps(cb, childProps, availableWidth, box.FontSize)

			cbWidth := availableWidth - cb.MarginLeft - cb.MarginRight -
				cb.PaddingLeft - cb.PaddingRight
			if mv, ok := childProps["max-width"]; ok {
				ctx := css.Context{FontSize: cb.FontSize, ViewportW: ViewportWidth, ViewportH: ViewportHeight, ParentWidth: availableWidth}
				maxW := css.Length(mv, ctx)
				if maxW > 0 && cbWidth > maxW {
					cbWidth = maxW
					if mv, ok := childProps["margin"]; ok && strings.Contains(mv, "auto") {
						offset := (availableWidth - cbWidth -
							cb.PaddingLeft - cb.PaddingRight) / 2
						cb.MarginLeft = offset
						cb.MarginRight = offset
					}
				}
			}

			cb.X = box.X + box.PaddingLeft + cb.MarginLeft + cb.PaddingLeft
			cb.Y = curY + cb.MarginTop + cb.PaddingTop

			layoutChildren(cb, rules, childProps, cbWidth, childHref)

			cb.Height = measureInnerHeight(cb)

			box.Children = append(box.Children, cb)
			curY += cb.MarginTop + cb.PaddingTop + cb.Height + cb.PaddingBottom + cb.MarginBottom
		} else {
			inlineRun = append(inlineRun, child)
			inlineProps = append(inlineProps, childProps)
		}
	}
	flushInline()

	box.Height = curY - box.Y + box.PaddingBottom
}

// measureInnerHeight — высота контента блока = максимальная нижняя граница детей.
func measureInnerHeight(b *Box) float64 {
	if len(b.Children) == 0 {
		return 0
	}
	maxY := b.Y
	for _, c := range b.Children {
		if end := c.Y + c.Height; end > maxY {
			maxY = end
		}
	}
	return maxY - b.Y
}

// fmtSscan — обёртка над fmt.Sscan.
func fmtSscan(s string, f *float64) (int, error) {
	return sscan(s, f)
}