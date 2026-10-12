package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// strokeGeometry builds the centerline path of an OpStrokeRect and its width
// in canvas pixels. A non-zero StrokeMask builds straight side segments; an
// unknown mask bit returns ok false. Zero builds a rounded path whose corners
// may be elliptical.
func strokeGeometry(op *layout.DisplayOp, dx, dy float64) (vector.Path, float32, bool) {
	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}

	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)

	if op.StrokeMask != 0 {
		return maskPathAt(x, y, w, h, op.StrokeMask), width, op.StrokeMask&^0x0F == 0
	}

	rx, ry := render.RadiiXY(op)

	return roundedPathAt(x, y, w, h, radius4(rx), radius4(ry)), width, true
}

// strokePathAt traces the centerline a raster key names.
func strokePathAt(key rasterKey) vector.Path {
	if key.mask != 0 {
		return maskPathAt(key.fx, key.fy, key.w, key.h, key.mask)
	}

	return roundedPathAt(key.fx, key.fy, key.w, key.h, key.r, key.ry)
}

// maskPathAt traces the selected sides as straight segments.
func maskPathAt(x, y, w, h float32, mask uint8) vector.Path {
	var path vector.Path

	if mask&1 != 0 {
		path.MoveTo(x, y)
		path.LineTo(x+w, y)
	}
	if mask&2 != 0 {
		path.MoveTo(x+w, y)
		path.LineTo(x+w, y+h)
	}
	if mask&4 != 0 {
		path.MoveTo(x, y+h)
		path.LineTo(x+w, y+h)
	}
	if mask&8 != 0 {
		path.MoveTo(x, y)
		path.LineTo(x, y+h)
	}

	return path
}
