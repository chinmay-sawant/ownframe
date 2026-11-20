package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"math"
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

// quarterTurn is a quarter circle in radians, one rounded-corner arc span.
const quarterTurn = math.Pi / 2

// threeQuarterTurn is three quarters of a circle, the last corner arc end.
const threeQuarterTurn = 3 * quarterTurn

// populateRoundedPath traces a border centerline with per-corner elliptical arcs
// from its origin into path.
func populateRoundedPath(path *vector.Path, x, y, w, h float32, rx, ry [4]float32) {
	path.MoveTo(x+rx[0], y)
	path.LineTo(x+w-rx[1], y)
	ellipseTo(path, x+w-rx[1], y+ry[1], rx[1], ry[1], -quarterTurn, 0)
	path.LineTo(x+w, y+h-ry[2])
	ellipseTo(path, x+w-rx[2], y+h-ry[2], rx[2], ry[2], 0, quarterTurn)
	path.LineTo(x+rx[3], y+h)
	ellipseTo(path, x+rx[3], y+h-ry[3], rx[3], ry[3], quarterTurn, math.Pi)
	path.LineTo(x, y+ry[0])
	ellipseTo(path, x+rx[0], y+ry[0], rx[0], ry[0], math.Pi, threeQuarterTurn)
	path.Close()
}

// ellipseTo appends an elliptical arc as short segments. The path must
// already sit on the arc's start point.
func ellipseTo(path *vector.Path, cx, cy, rx, ry, from, to float32) {
	const steps = 8

	for i := 1; i <= steps; i++ {
		a := from + (to-from)*float32(i)/steps
		path.LineTo(cx+rx*float32(math.Cos(float64(a))), cy+ry*float32(math.Sin(float64(a))))
	}
}
