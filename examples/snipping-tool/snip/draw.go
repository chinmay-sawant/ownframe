package snip

import (
	"image"
	"image/color"
	"math"
)

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// dot sets one pixel when inside dst.
func dot(dst *image.RGBA, x, y int, c color.Color) {
	if (image.Point{X: x, Y: y}).In(dst.Bounds()) {
		dst.Set(x, y, c)
	}
}

// line draws a 1px Bresenham segment.
func line(dst *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx, dy := abs(x1-x0), -abs(y1-y0)
	sx, sy := 1, 1
	if x0 > x1 {
		sx = -1
	}
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		dot(dst, x0, y0, c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e := bresenhamFactor * err
		if e >= dy {
			err += dy
			x0 += sx
		}
		if e <= dx {
			err += dx
			y0 += sy
		}
	}
}

// head draws two barbs behind the tip (x1, y1).
func head(dst *image.RGBA, x0, y0, x1, y1 int, c color.Color) {
	dx, dy := float64(x1-x0), float64(y1-y0)
	if n := math.Hypot(dx, dy); n > 0 {
		ux, uy := dx/n, dy/n
		ax, ay := float64(x1)-ux*arrowLen, float64(y1)-uy*arrowLen
		px, py := -uy*arrowWidth, ux*arrowWidth
		line(dst, x1, y1, int(ax+px), int(ay+py), c)
		line(dst, x1, y1, int(ax-px), int(ay-py), c)
	}
}

// rectOutline draws a 2px box outline.
func rectOutline(dst *image.RGBA, x, y, w, h int, c color.Color) {
	for t := 0; t < 2; t++ {
		for i := x; i < x+w; i++ {
			dot(dst, i, y+t, c)
			dot(dst, i, y+h-1-t, c)
		}
		for j := y; j < y+h; j++ {
			dot(dst, x+t, j, c)
			dot(dst, x+w-1-t, j, c)
		}
	}
}

// textStrip fills a 3px strip inside the top edge as the label
// placeholder; no glyphs are drawn without a font engine.
func textStrip(dst *image.RGBA, x, y, w int, c color.Color) {
	for j := y + stripInset; j < y+5; j++ {
		for i := x + stripInset; i < x+w-stripInset; i++ {
			dot(dst, i, j, c)
		}
	}
}
