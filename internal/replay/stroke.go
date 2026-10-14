package replay

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

// drawStrokeRect strokes a rounded rectangle. A zero StrokeMask strokes the
// whole border; a non-zero mask strokes only the selected sides. Corners may
// be elliptical. A repeated shape comes from the raster cache.
func drawStrokeRect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if cachedStroke(dst, op, dx, dy) {
		return
	}

	strokeDirect(dst, op, dx, dy)
}

// strokeDirect strokes the outline without the raster cache. An unknown mask
// bit paints nothing.
func strokeDirect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	path, width, ok := strokeGeometry(op, dx, dy)
	if !ok {
		return
	}

	stroke := &vector.StrokeOptions{Width: width, LineJoin: vector.LineJoinRound}
	options := &vector.DrawPathOptions{AntiAlias: true}
	options.ColorScale.ScaleWithColor(rgba(op))

	vector.StrokePath(dst, &path, stroke, options)
}
