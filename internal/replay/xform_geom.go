package replay

import (
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
	g.SetElement(0, geomTransCol, pxPerPt*(m.A*op.X+m.C*op.Y+m.E)+dx)
	g.SetElement(1, 0, m.B)
	g.SetElement(1, 1, m.D)
	g.SetElement(1, geomTransCol, pxPerPt*(m.B*op.X+m.D*op.Y+m.F)+dy)

	return g
}
