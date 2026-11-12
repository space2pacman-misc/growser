package colorutil

import (
	"image/color"
	"strconv"
	"strings"
)

// Parse поддерживает #rgb, #rrggbb и несколько именованных.
func Parse(s string) (color.RGBA, bool) {
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