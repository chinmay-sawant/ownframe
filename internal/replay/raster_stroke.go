package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// cachedStroke draws a rounded or masked border through the raster cache. ok
// is false for a shape too large to cache or an unknown mask bit, which the
// caller handles as before.
func (renderer *Renderer) cachedStroke(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) bool {
	if op.StrokeMask&^0x0F != 0 {
		return false
	}

	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}

	left, top, size, ok := rasterBox(x, y, w, h, width/2+1)
	if !ok {
		return false
	}

	rx, ry := masking.RadiiXY(op)
	key := rasterKey{
		variant: rasterStroke,
		mask:    op.StrokeMask,
		w:       w, h: h, width: width,
		r:  radius4(rx),
		ry: radius4(ry),
		fx: x - float32(left),
		fy: y - float32(top),
	}
	img := renderer.rasters.raster(key, size, func(img *ebiten.Image) {
		path := strokePathAt(key)
		stroke := &vector.StrokeOptions{Width: width, LineJoin: vector.LineJoinRound}
		options := &vector.DrawPathOptions{AntiAlias: true}
		vector.StrokePath(img, &path, stroke, options)
	})
	renderer.blitRaster(dst, img, left, top, rgba(op))

	return true
}
