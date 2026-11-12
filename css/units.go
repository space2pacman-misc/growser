package css

import (
	"strconv"
	"strings"
)

// Context — размеры для вычисления em/vh/%.
type Context struct {
	FontSize      float64 // текущий font-size, px
	ViewportW     float64
	ViewportH     float64
	ParentWidth   float64 // для %
}

// Length возвращает длину в пикселях.
func Length(val string, ctx Context) float64 {
	val = strings.TrimSpace(strings.ToLower(val))
	if val == "" || val == "auto" {
		return 0
	}

	// vh / vw / em / % / px
	for _, unit := range []string{"vh", "vw", "em", "%", "px"} {
		if strings.HasSuffix(val, unit) {
			numStr := strings.TrimSuffix(val, unit)
			n, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0
			}
			switch unit {
			case "px":
				return n
			case "em":
				return n * ctx.FontSize
			case "vh":
				return n * ctx.ViewportH / 100
			case "vw":
				return n * ctx.ViewportW / 100
			case "%":
				return n * ctx.ParentWidth / 100
			}
		}
	}

	// без единиц — считаем px
	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0
	}
	return n
}