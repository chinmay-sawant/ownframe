package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawStrokeRect strokes a rounded rectangle. A zero StrokeMask strokes the
// whole border; a non-zero mask strokes only the selected sides. Corners may
// be elliptical. A repeated shape comes from the raster cache.
func (renderer *Renderer) drawStrokeRect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if renderer.cachedStroke(dst, op, dx, dy) {
		return
	}

	strokeDirect(dst, op, dx, dy)
}

// strokeDirect strokes the outline without the raster cache. An unknown mask
// bit paints nothing.
func strokeDirect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	path := acquireVectorPath()
	width, ok := populateStrokeGeometry(path, op, dx, dy)
	if !ok {
		releaseVectorPath(path)
		return
	}

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineJoin = vector.LineJoinRound

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/colorChannelMax, float32(cg)/colorChannelMax, float32(cb)/colorChannelMax, float32(ca)/colorChannelMax)

	vector.StrokePath(dst, path, &stroke, &options)
	releaseVectorPath(path)
}
