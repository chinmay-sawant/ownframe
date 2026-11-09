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
	path := acquireVectorPath()
	path.MoveTo(float32(x0*pxPerPt+dx), float32(y0*pxPerPt+dy))
	path.LineTo(float32(x1*pxPerPt+dx), float32(y1*pxPerPt+dy))

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineCap = vector.LineCapSquare

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/colorChannelMax, float32(cg)/colorChannelMax, float32(cb)/colorChannelMax, float32(ca)/colorChannelMax)

	vector.StrokePath(dst, path, &stroke, &options)
	releaseVectorPath(path)
}
