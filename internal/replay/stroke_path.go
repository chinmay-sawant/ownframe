package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// populateStrokeGeometry traces the centerline path of an OpStrokeRect into path
// and computes its width in canvas pixels. A non-zero StrokeMask builds straight
// side segments; an unknown mask bit returns ok false. Zero builds a rounded path
// whose corners may be elliptical.
func populateStrokeGeometry(path *vector.Path, op *layout.DisplayOp, dx, dy float64) (float32, bool) {
	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}

	if op.StrokeMask&^0x0F != 0 {
		return 0, false
	}

	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)

	if op.StrokeMask != 0 {
		populateMaskPath(path, x, y, w, h, op.StrokeMask)
		return width, true
	}

	rx, ry := masking.RadiiXY(op)
	populateRoundedPath(path, x, y, w, h, radius4(rx), radius4(ry))

	return width, true
}

// strokeGeometry builds the centerline path of an OpStrokeRect and its width
// in canvas pixels. A non-zero StrokeMask builds straight side segments; an
// unknown mask bit returns ok false. Zero builds a rounded path whose corners
// may be elliptical.
func strokeGeometry(op *layout.DisplayOp, dx, dy float64) (vector.Path, float32, bool) {
	var path vector.Path
	width, ok := populateStrokeGeometry(&path, op, dx, dy)
	return path, width, ok
}

// populateStrokePath traces the centerline a raster key names into path.
func populateStrokePath(path *vector.Path, key rasterKey) {
	if key.mask != 0 {
		populateMaskPath(path, key.fx, key.fy, key.w, key.h, key.mask)
		return
	}

	populateRoundedPath(path, key.fx, key.fy, key.w, key.h, key.r, key.ry)
}

// strokePathAt traces the centerline a raster key names.
func strokePathAt(key rasterKey) vector.Path {
	var path vector.Path
	populateStrokePath(&path, key)
	return path
}

// populateMaskPath traces the selected sides as straight segments into path.
func populateMaskPath(path *vector.Path, x, y, w, h float32, mask uint8) {
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
}

// maskPathAt traces the selected sides as straight segments.
func maskPathAt(x, y, w, h float32, mask uint8) vector.Path {
	var path vector.Path
	populateMaskPath(&path, x, y, w, h, mask)
	return path
}
