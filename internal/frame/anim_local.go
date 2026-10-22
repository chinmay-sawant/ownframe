package frame

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// localBox is op's painted box in points before its transform.
// A nil op reports the zero box.
func localBox(op *page.DisplayOp) (x0, y0, x1, y1 float64) {
	if op == nil {
		return 0, 0, 0, 0
	}
	x0, y0 = op.X, op.Y
	x1, y1 = op.X+op.W, op.Y+op.H

	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		y0, y1 = op.Y-ascent(op), op.Y+op.InkDescent
	case layout.DisplayOpStrokeRect, layout.DisplayOpLine:
		g := math.Max(op.Width, 1) / halfDivisor
		x0, y0, x1, y1 = x0-g, y0-g, x1+g, y1+g
	}

	return x0, y0, x1, y1
}

// ascent is the face ascent in points for one text op.
func ascent(op *page.DisplayOp) float64 {
	if op.Font == nil || op.Font.UnitsPerEm() <= 0 {
		return op.Size
	}

	return float64(op.Font.Ascent()) / float64(op.Font.UnitsPerEm()) * op.Size
}

// paintCenter is op's painted center in points, the origin scale and rotate
// turn about.
func paintCenter(op *page.DisplayOp) (float64, float64) {
	x0, y0, x1, y1 := localBox(op)

	return (x0 + x1) / halfDivisor, (y0 + y1) / halfDivisor
}
