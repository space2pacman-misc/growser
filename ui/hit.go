package ui

import "growser/layout"

// HitTest возвращает самый глубокий Box, содержащий точку (x, y) в координатах страницы.
func HitTest(b *layout.Box, x, y float64) *layout.Box {
	x1 := b.X - b.PaddingLeft
	y1 := b.Y - b.PaddingTop
	x2 := b.X + b.Width + b.PaddingRight
	y2 := b.Y + b.Height + b.PaddingBottom
	if x < x1 || x > x2 || y < y1 || y > y2 {
		return nil
	}
	for _, c := range b.Children {
		if hit := HitTest(c, x, y); hit != nil {
			return hit
		}
	}
	return b
}