package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"image/color"
	"math"
)

// drawStrokeRect strokes a rounded rectangle. A zero StrokeMask strokes the
// whole border; a non-zero mask strokes only the selected sides. Corners may
// be elliptical. A repeated shape comes from the raster cache.
func (renderer *Renderer) drawStrokeRect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if renderer.cachedStroke(dst, op, dx, dy) {
		return
	}

	renderer.strokeDirect(dst, op, dx, dy)
}

// strokeDirect strokes the outline without the raster cache using the default renderer.
func strokeDirect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	defaultRenderer.strokeDirect(dst, op, dx, dy)
}

// strokeDirect strokes the outline without the raster cache. An unknown mask
// bit paints nothing.
func (renderer *Renderer) strokeDirect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.StrokeMask&^0x0F != 0 {
		return
	}
	width := float32(op.Width * pxPerPt)
	if width < 1 {
		width = 1
	}
	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	rx, ry := masking.RadiiXY(op)
	renderer.strokeRectNineSlice(dst, x, y, w, h, width, op.StrokeMask, radius4(rx), radius4(ry), rgba(op))
}

// strokeRectNineSlice paints an axis-aligned rectangle border by splitting it
// into up to four straight filled edges and up to four small corner arcs culled
// to dst.Bounds(). Small corners use cached analytical masks instead of
// Ebiten's vector.StrokePath, keeping the permanent graphics command queue tiny
// while preserving anti-aliased rounded corners.
func (renderer *Renderer) strokeRectNineSlice(
	dst *ebiten.Image,
	x, y, w, h, width float32,
	mask uint8,
	rx, ry [4]float32,
	clr color.RGBA,
) {
	if dst == nil || w <= 0 || h <= 0 || width <= 0 || clr.A == 0 {
		return
	}
	if mask != 0 {
		strokeMaskedSides(dst, x, y, w, h, width, mask, clr)
		return
	}

	half := width / halfDiv
	for i := range 4 {
		if rx[i] <= 0 || ry[i] <= 0 {
			rx[i], ry[i] = 0, 0
		}
	}

	// Top straight edge.
	topX0 := x + rx[0]
	if rx[0] == 0 {
		topX0 = x - half
	}
	topX1 := x + w - rx[1]
	if rx[1] == 0 {
		topX1 = x + w + half
	}
	fillEdgeClipped(dst, topX0, y-half, topX1-topX0, width, clr)

	// Bottom straight edge.
	botX0 := x + rx[3]
	if rx[3] == 0 {
		botX0 = x - half
	}
	botX1 := x + w - rx[2]
	if rx[2] == 0 {
		botX1 = x + w + half
	}
	fillEdgeClipped(dst, botX0, y+h-half, botX1-botX0, width, clr)

	// Right straight edge.
	rightY0 := y + ry[1]
	if ry[1] == 0 {
		rightY0 = y + half
	}
	rightY1 := y + h - ry[2]
	if ry[2] == 0 {
		rightY1 = y + h - half
	}
	fillEdgeClipped(dst, x+w-half, rightY0, width, rightY1-rightY0, clr)

	// Left straight edge.
	leftY0 := y + ry[0]
	if ry[0] == 0 {
		leftY0 = y + half
	}
	leftY1 := y + h - ry[3]
	if ry[3] == 0 {
		leftY1 = y + h - half
	}
	fillEdgeClipped(dst, x-half, leftY0, width, leftY1-leftY0, clr)

	bounds := dst.Bounds()
	// Top-right corner: -pi/2 -> 0.
	if rx[1] > 0 {
		renderer.strokeCorner(dst, bounds, x+w-rx[1], y+ry[1], rx[1], ry[1], -quarterTurn, 0, width, cornerTR, clr)
	}
	// Bottom-right corner: 0 -> pi/2.
	if rx[2] > 0 {
		renderer.strokeCorner(dst, bounds, x+w-rx[2], y+h-ry[2], rx[2], ry[2], 0, quarterTurn, width, cornerBR, clr)
	}
	// Bottom-left corner: pi/2 -> pi.
	if rx[3] > 0 {
		renderer.strokeCorner(dst, bounds, x+rx[3], y+h-ry[3], rx[3], ry[3], quarterTurn, math.Pi, width, cornerBL, clr)
	}
	// Top-left corner: pi -> 3*pi/2.
	if rx[0] > 0 {
		renderer.strokeCorner(dst, bounds, x+rx[0], y+ry[0], rx[0], ry[0], math.Pi, threeQuarterTurn, width, cornerTL, clr)
	}
}

// Border stroke mask bit flags.
const (
	strokeMaskTop    uint8 = 1 << 0
	strokeMaskRight  uint8 = 1 << 1
	strokeMaskBottom uint8 = 1 << 2
	strokeMaskLeft   uint8 = 1 << 3
)

// strokeMaskedSides draws non-overlapping axis-aligned border sides selected by
// mask bits (1=top, 2=right, 4=bottom, 8=left).
func strokeMaskedSides(dst *ebiten.Image, x, y, w, h, width float32, mask uint8, clr color.RGBA) {
	half := width / halfDiv
	hasTop := mask&strokeMaskTop != 0
	hasRight := mask&strokeMaskRight != 0
	hasBottom := mask&strokeMaskBottom != 0
	hasLeft := mask&strokeMaskLeft != 0

	if hasTop {
		x0, x1 := x, x+w
		if hasLeft {
			x0 -= half
		}
		if hasRight {
			x1 += half
		}
		fillEdgeClipped(dst, x0, y-half, x1-x0, width, clr)
	}
	if hasBottom {
		x0, x1 := x, x+w
		if hasLeft {
			x0 -= half
		}
		if hasRight {
			x1 += half
		}
		fillEdgeClipped(dst, x0, y+h-half, x1-x0, width, clr)
	}
	if hasRight {
		y0, y1 := y, y+h
		if hasTop {
			y0 += half
		}
		if hasBottom {
			y1 -= half
		}
		fillEdgeClipped(dst, x+w-half, y0, width, y1-y0, clr)
	}
	if hasLeft {
		y0, y1 := y, y+h
		if hasTop {
			y0 += half
		}
		if hasBottom {
			y1 -= half
		}
		fillEdgeClipped(dst, x-half, y0, width, y1-y0, clr)
	}
}

// fillEdgeClipped fills one axis-aligned rectangle clipped to dst.Bounds().
func fillEdgeClipped(dst *ebiten.Image, x, y, w, h float32, clr color.RGBA) {
	if w <= 0 || h <= 0 {
		return
	}
	b := dst.Bounds()
	minX := max(x, float32(b.Min.X))
	minY := max(y, float32(b.Min.Y))
	maxX := min(x+w, float32(b.Max.X))
	maxY := min(y+h, float32(b.Max.Y))
	if maxX <= minX || maxY <= minY {
		return
	}
	vector.FillRect(dst, minX, minY, maxX-minX, maxY-minY, clr, false)
}

// strokeCorner paints one quarter-ellipse corner arc using a cached analytical
// mask when within cache limits, falling back to a clipped vector stroke arc.
func (renderer *Renderer) strokeCorner(
	dst *ebiten.Image,
	dstBounds image.Rectangle,
	cx, cy, rx, ry, from, to, width float32,
	corner roundedCorner,
	clr color.RGBA,
) {
	if renderer.strokeCornerAnalytical(dst, dstBounds, cx, cy, rx, ry, width, corner, clr) {
		return
	}
	strokeCornerArc(dst, dstBounds, cx, cy, rx, ry, from, to, width, clr)
}

// strokeCornerArc strokes a single quarter-ellipse corner arc when its bounds
// intersect dstBounds.
func strokeCornerArc(
	dst *ebiten.Image,
	dstBounds image.Rectangle,
	cx, cy, rx, ry, from, to, width float32,
	clr color.RGBA,
) {
	pad := width/halfDiv + rasterPad
	box := image.Rect(
		int(math.Floor(float64(cx-rx-pad))),
		int(math.Floor(float64(cy-ry-pad))),
		int(math.Ceil(float64(cx+rx+pad))),
		int(math.Ceil(float64(cy+ry+pad))),
	)
	if !box.Overlaps(dstBounds) {
		return
	}

	path := acquireVectorPath()
	path.MoveTo(cx+rx*float32(math.Cos(float64(from))), cy+ry*float32(math.Sin(float64(from))))
	ellipseTo(path, cx, cy, rx, ry, from, to)

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineJoin = vector.LineJoinRound

	var options vector.DrawPathOptions
	options.AntiAlias = true
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(
		float32(cr)/colorChannelMax,
		float32(cg)/colorChannelMax,
		float32(cb)/colorChannelMax,
		float32(ca)/colorChannelMax,
	)

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
