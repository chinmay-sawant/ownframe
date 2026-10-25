package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// xformBounds bounds a transformed op in CSS pixels: the corners of its local
// box map through the matrix, so a rotation or skew bounds as the mapped
// parallelogram instead of its layout box. ok is false for a kind the
// painter cannot bound, such as a transformed grid.
func xformBounds(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	return matrixBounds(op, ppt, op.Transform())
}

func matrixBounds(op *layout.DisplayOp, ppt float64, m layout.Matrix2D) (image.Rectangle, bool) {
	l, t, r, b, ok := xformPointsBox(op)
	if !ok {
		return image.Rectangle{}, false
	}

	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{l, t}, {r, t}, {r, b}, {l, b}} {
		x, y := m.Apply(p[0], p[1])
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	pad := float64(boundsPad)

	return image.Rect(
		int(math.Floor(loX/ppt-pad)), int(math.Floor(loY/ppt-pad)),
		int(math.Ceil(hiX/ppt+pad)), int(math.Ceil(hiY/ppt+pad)),
	), true
}

// xformPointsBox is the op's local painted box in points, before its baked
// transform: the same geometry the untransformed bounds cover.
func xformPointsBox(op *layout.DisplayOp) (l, t, r, b float64, ok bool) {
	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		top := op.Y - textAscent(op)
		bottom := op.Y + op.InkDescent
		if bottom < top {
			bottom = top
		}

		return op.X, top, op.X + op.W, bottom, true
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1) / halfDiv

		return min(x, x+w) - grow, min(y, y+h) - grow,
			max(x, x+w) + grow, max(y, y+h) + grow, true
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1) / halfDiv

		return op.X - grow, op.Y - grow,
			op.X + op.W + grow, op.Y + op.H + grow, true
	case layout.DisplayOpFillRect:
		return op.X, op.Y, op.X + op.W, op.Y + op.H, true
	}

	return 0, 0, 0, 0, false
}
