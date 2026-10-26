package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// drawXformScratch rasterizes a shape too large for the raster cache into a
// local scratch image and blits it through the op matrix.
func (renderer *Renderer) drawXformScratch(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	left, top, size, ok := xformScratch(op)
	if !ok {
		return
	}

	scratch := ebiten.NewImage(size.X, size.Y)
	defer scratch.Deallocate()

	ox, oy := float64(-left), float64(-top)

	switch op.Kind {
	case layout.DisplayOpFillRect:
		renderer.fillRect(scratch, op, ox, oy)
	case layout.DisplayOpStrokeRect:
		renderer.drawStrokeRect(scratch, op, ox, oy)
	case layout.DisplayOpLine:
		drawLine(scratch, op, ox, oy)
	}

	// Blit the interior: a bilinear tap at a whole image's boundary reads a
	// texel outside its region, whose content is undefined, and the shape
	// keeps a two-pixel pad, so one pixel of inset costs no ink.
	src := scratch
	if size.X > 2 && size.Y > 2 {
		if inner, ok := scratch.SubImage(image.Rect(1, 1, size.X-1, size.Y-1)).(*ebiten.Image); ok {
			src = inner
			left++
			top++
		}
	}

	geom := xformGeom(op, dx, dy)
	shiftLocal(&geom, float64(left)-op.X*pxPerPt, float64(top)-op.Y*pxPerPt)
	clipped, paint := scissorOverflow(dst, op, dx, dy)
	if !paint {
		return
	}
	clipped.DrawImage(src, &ebiten.DrawImageOptions{
		GeoM:   geom,
		Filter: ebiten.FilterLinear,
	})
}
