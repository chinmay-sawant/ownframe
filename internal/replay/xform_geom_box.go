package replay

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// shiftLocal moves a geom's origin by (x, y) in its local space, before the
// matrix. Ebiten's GeoM.Translate moves the origin in dst space instead.
func shiftLocal(g *ebiten.GeoM, x, y float64) {
	g.SetElement(0, geomTransCol, g.Element(0, geomTransCol)+g.Element(0, 0)*x+g.Element(0, 1)*y)
	g.SetElement(1, geomTransCol, g.Element(1, geomTransCol)+g.Element(1, 0)*x+g.Element(1, 1)*y)
}

// xformLocalBox is an op's painted box in its own device pixels, before the
// baked transform: the geometry the untransformed painter would draw.
func xformLocalBox(op *layout.DisplayOp) (l, t, r, b float64, ok bool) {
	k := pxPerPt

	switch op.Kind {
	case layout.DisplayOpFillRect:
		return op.X * k, op.Y * k, (op.X + op.W) * k, (op.Y + op.H) * k, true
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1) / halfDiv * k

		return op.X*k - grow, op.Y*k - grow,
			(op.X+op.W)*k + grow, (op.Y+op.H)*k + grow, true
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1) / halfDiv * k
		x0, x1 := min(x, x+w), max(x, x+w)
		y0, y1 := min(y, y+h), max(y, y+h)

		return x0*k - grow, y0*k - grow, x1*k + grow, y1*k + grow, true
	}

	return 0, 0, 0, 0, false
}
