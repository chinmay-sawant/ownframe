package window

import (
	"image/color"

	"github.com/chinmay-sawant/blinkless/layout"
)

// pageBackground returns the color to paint under the page: the page's own
// background fill for a display list, or the corner pixel of the bitmap.
// White when neither is known.
func (s *shell) pageBackground() color.Color {
	if s.transparent {
		return color.Transparent
	}

	if s.display != nil {
		if c, ok := displayBackground(s.display); ok {
			return c
		}
	}

	if s.img != nil {
		b := s.img.Bounds()
		if !b.Empty() {
			return s.img.At(b.Min.X, b.Min.Y)
		}
	}

	return color.White
}

// displayBackground finds the first fill in paint order that starts at the
// canvas top-left and spans its width: the html or body background.
func displayBackground(display *layout.Display) (color.Color, bool) {
	if display.PointsPerPixel <= 0 || display.Width <= 0 {
		return nil, false
	}

	widthPts := float64(display.Width) * display.PointsPerPixel
	order := display.Order
	n := len(order)
	if n == 0 {
		n = len(display.Ops)
	}

	for k := 0; k < n; k++ {
		index := k
		if len(order) != 0 {
			index = order[k]
		}

		if index < 0 || index >= len(display.Ops) {
			continue
		}

		op := &display.Ops[index]
		if op.Kind != layout.DisplayOpFillRect || op.X > 0.5 || op.Y > 0.5 {
			continue
		}

		if op.X+op.W < widthPts-0.5 {
			continue
		}

		return opColor(op), true
	}

	return nil, false
}

// opColor returns the op's color composited over white.
func opColor(op *layout.DisplayOp) color.Color {
	a := op.Alpha
	if a <= 0 || a > 1 {
		a = 1
	}

	return color.RGBA{
		R: uint8(clamp01(op.R*a+1-a) * 255),
		G: uint8(clamp01(op.G*a+1-a) * 255),
		B: uint8(clamp01(op.B*a+1-a) * 255),
		A: 255,
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}

	if v > 1 {
		return 1
	}

	return v
}
