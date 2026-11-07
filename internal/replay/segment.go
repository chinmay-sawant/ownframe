package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawSegment strokes a diagonal line drawn in point space. Square caps
// extend the stroke half its width past each endpoint, matching the engine.
func drawSegment(
	dst *ebiten.Image, op *layout.DisplayOp,
	x0, y0, x1, y1 float64, width float32, dx, dy float64,
) {
	path := vectorPathPool.Get().(*vector.Path)
	path.Reset()
	path.MoveTo(float32(x0*pxPerPt+dx), float32(y0*pxPerPt+dy))
	path.LineTo(float32(x1*pxPerPt+dx), float32(y1*pxPerPt+dy))

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineCap = vector.LineCapSquare

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/0xffff, float32(cg)/0xffff, float32(cb)/0xffff, float32(ca)/0xffff)

	vector.StrokePath(dst, path, &stroke, &options)
	path.Reset()
	vectorPathPool.Put(path)
}
