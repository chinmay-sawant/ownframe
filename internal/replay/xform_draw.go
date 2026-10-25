package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// drawXform draws one op through its baked transform: a shape comes from the
// raster cache and blits through the matrix, and a run of glyphs draws with
// the matrix as its GeoM. A transformed grid carries no paint, because the
// gate keeps it on the bitmap path.
func (renderer *Renderer) drawXform(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		renderer.drawXformText(dst, op, dx, dy)
	case layout.DisplayOpGridRun:
	default:
		renderer.drawXformShape(dst, op, dx, dy)
	}
}

// drawXformShape draws a fill, stroke, or line through the op matrix. The
// cached raster is reused across transforms; a shape too large to cache
// rasterizes into a scratch image instead.
func (renderer *Renderer) drawXformShape(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	img, at, ok := renderer.xformRaster(op)
	if !ok {
		renderer.drawXformScratch(dst, op, dx, dy)

		return
	}

	geom := xformGeom(op, dx, dy)
	shiftLocal(&geom, float64(at.X)-op.X*pxPerPt, float64(at.Y)-op.Y*pxPerPt)
	opts := &ebiten.DrawImageOptions{
		GeoM:   geom,
		Filter: ebiten.FilterLinear,
	}
	opts.ColorScale.ScaleWithColor(rgba(op))
	dst.DrawImage(img, opts)
}
