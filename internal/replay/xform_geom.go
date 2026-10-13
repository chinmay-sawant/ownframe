package replay

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// xformPad is the pixel slack around a transformed shape's local raster.
const xformPad = 2

// xformGeom maps an op's local device pixels through its baked transform into
// dst pixels. Local coordinates count from the op's own origin in CSS pixels;
// the matrix itself is unitless, so only its translation scales from points.
func xformGeom(op *layout.DisplayOp, dx, dy float64) ebiten.GeoM {
	m := op.Transform()

	var g ebiten.GeoM
	g.SetElement(0, 0, m.A)
	g.SetElement(0, 1, m.C)
	g.SetElement(0, 2, pxPerPt*(m.A*op.X+m.C*op.Y+m.E)+dx)
	g.SetElement(1, 0, m.B)
	g.SetElement(1, 1, m.D)
	g.SetElement(1, 2, pxPerPt*(m.B*op.X+m.D*op.Y+m.F)+dy)

	return g
}

// shiftLocal moves a geom's origin by (x, y) in its local space, before the
// matrix. Ebiten's GeoM.Translate moves the origin in dst space instead.
func shiftLocal(g *ebiten.GeoM, x, y float64) {
	g.SetElement(0, 2, g.Element(0, 2)+g.Element(0, 0)*x+g.Element(0, 1)*y)
	g.SetElement(1, 2, g.Element(1, 2)+g.Element(1, 0)*x+g.Element(1, 1)*y)
}

// xformLocalBox is an op's painted box in its own device pixels, before the
// baked transform: the geometry the untransformed painter would draw.
func xformLocalBox(op *layout.DisplayOp) (l, t, r, b float64, ok bool) {
	k := pxPerPt

	switch op.Kind {
	case layout.DisplayOpFillRect:
		return op.X * k, op.Y * k, (op.X + op.W) * k, (op.Y + op.H) * k, true
	case layout.DisplayOpStrokeRect:
		grow := math.Max(op.Width, 1) / 2 * k

		return op.X*k - grow, op.Y*k - grow,
			(op.X+op.W)*k + grow, (op.Y+op.H)*k + grow, true
	case layout.DisplayOpLine:
		x, y, w, h, width := op.PaintLineGeometry()
		grow := math.Max(width, 1) / 2 * k
		x0, x1 := min(x, x+w), max(x, x+w)
		y0, y1 := min(y, y+h), max(y, y+h)

		return x0*k - grow, y0*k - grow, x1*k + grow, y1*k + grow, true
	}

	return 0, 0, 0, 0, false
}
