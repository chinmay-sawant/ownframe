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
	path := vectorPathPool.Get().(*vector.Path)
	path.Reset()
	width, ok := populateStrokeGeometry(path, op, dx, dy)
	if !ok {
		path.Reset()
		vectorPathPool.Put(path)
		return
	}

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineJoin = vector.LineJoinRound

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/0xffff, float32(cg)/0xffff, float32(cb)/0xffff, float32(ca)/0xffff)

	vector.StrokePath(dst, path, &stroke, &options)
	path.Reset()
	vectorPathPool.Put(path)
}
