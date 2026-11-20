package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image"
	"math"
	"sync"
)

// fillRect draws one filled rectangle. Corners are circular; the caller has
// already rejected elliptical corners. Plain fills snap to whole pixels, the
// same geometry the engine's raster painter uses.
func (renderer *Renderer) fillRect(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	radii, ok := masking.FillRadii(op)
	if !ok {
		return
	}

	if radii == ([4]float64{}) {
		left, top, right, bottom := snapped(op, dx, dy)
		vector.FillRect(dst, left, top, right-left, bottom-top, rgba(op), false)

		return
	}

	// A repeated rounded shape comes from the raster cache; a shape too large
	// to cache falls through to the direct path.
	if renderer.cachedRoundedFill(dst, op, radii, dx, dy) {
		return
	}

	fillRounded(dst, op, radii, dx, dy)
}

// snapped converts an op's rectangle to whole canvas pixels, matching the
// engine's ptRectScale rounding.
func snapped(op *layout.DisplayOp, dx, dy float64) (left, top, right, bottom float32) {
	return float32(math.Round(op.X*pxPerPt + dx)), float32(math.Round(op.Y*pxPerPt + dy)),
		float32(math.Round((op.X+op.W)*pxPerPt + dx)), float32(math.Round((op.Y+op.H)*pxPerPt + dy))
}

// vectorPathPool holds vector paths for the direct painters, so a repeated
// shape reuses its path slices after the first paint.
var vectorPathPool = sync.Pool{ //nolint:gochecknoglobals // reason: process-wide pool that keeps every direct painter allocation-free after warm-up.
	New: func() any { return new(vector.Path) },
}

// acquireVectorPath borrows a path from the shared pool and resets it.
func acquireVectorPath() *vector.Path {
	path, ok := vectorPathPool.Get().(*vector.Path)
	if !ok || path == nil {
		return new(vector.Path)
	}

	path.Reset()

	return path
}

// releaseVectorPath resets path and returns it to the shared pool.
func releaseVectorPath(path *vector.Path) {
	path.Reset()
	vectorPathPool.Put(path)
}

// fillRounded draws a fill with circular corners as one vector path. The
// engine's painter keeps the real corners when the canvas clips the box, so
// the path is built from the op rectangle itself.
func fillRounded(dst *ebiten.Image, op *layout.DisplayOp, radii [4]float64, dx, dy float64) {
	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	r := radius4(radii)

	path := acquireVectorPath()
	populateRoundedFillPath(path, x, y, w, h, r)

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/colorChannelMax, float32(cg)/colorChannelMax, float32(cb)/colorChannelMax, float32(ca)/colorChannelMax)

	vector.FillPath(dst, path, nil, &options)
	releaseVectorPath(path)
}

// populateRoundedFillPath traces a fill's corners as circular arcs from its origin into path.
func populateRoundedFillPath(path *vector.Path, x, y, w, h float32, r [4]float32) {
	path.MoveTo(x+r[0], y)
	path.LineTo(x+w-r[1], y)
	arcTo(path, x+w, y, x+w, y+r[1], r[1])
	path.LineTo(x+w, y+h-r[2])
	arcTo(path, x+w, y+h, x+w-r[2], y+h, r[2])
	path.LineTo(x+r[3], y+h)
	arcTo(path, x, y+h, x, y+h-r[3], r[3])
	path.LineTo(x, y+r[0])
	arcTo(path, x, y, x+r[0], y, r[0])
	path.Close()
}

// radius4 scales a corner radius set from layout points to canvas pixels.
func radius4(radii [4]float64) [4]float32 {
	return [4]float32{
		float32(radii[0] * pxPerPt), float32(radii[1] * pxPerPt),
		float32(radii[2] * pxPerPt), float32(radii[3] * pxPerPt),
	}
}

func arcTo(path *vector.Path, x1, y1, x2, y2, radius float32) {
	if radius > 0 {
		path.ArcTo(x1, y1, x2, y2, radius)
	}
}

// lineHalf scales a stroke to its half-extent and back.
const lineHalf = 2

// drawLine draws one border segment. The engine centers the stroke on the
// line and extends it half a stroke past each endpoint, so one filled
// rectangle reproduces an axis-aligned segment exactly. A diagonal segment,
// such as a checked checkbox's tick, is stroked instead, with square caps
// for the same extension; a filled rectangle would flatten it into a bar.
func drawLine(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	x, y, w, h, width := op.PaintLineGeometry()

	stroke := int(math.Round(width * pxPerPt))
	if stroke < 1 {
		stroke = 1
	}

	if w != 0 && h != 0 {
		drawSegment(dst, op, x, y, x+w, y+h, float32(stroke), dx, dy)

		return
	}

	half := float64(stroke) / lineHalf / pxPerPt

	var box image.Rectangle
	if op.H <= 0 {
		box = snap(x-half, y-half, w+lineHalf*half, lineHalf*half)
	} else {
		box = snap(x-half, y-half, lineHalf*half, h+lineHalf*half)
	}

	vector.FillRect(
		dst,
		float32(box.Min.X)+float32(dx), float32(box.Min.Y)+float32(dy),
		float32(box.Dx()), float32(box.Dy()), rgba(op), false,
	)
}

// snap converts a point-space rectangle to whole canvas pixels, matching the
// engine's own raster geometry.
func snap(x, y, w, h float64) image.Rectangle {
	return image.Rect(
		int(math.Round(x*pxPerPt)), int(math.Round(y*pxPerPt)),
		int(math.Round((x+w)*pxPerPt)), int(math.Round((y+h)*pxPerPt)),
	)
}

// drawSegment strokes a diagonal line drawn in point space. Square caps
// extend the stroke half its width past each endpoint, matching the engine.
func drawSegment(
	dst *ebiten.Image, op *layout.DisplayOp,
	x0, y0, x1, y1 float64, width float32, dx, dy float64,
) {
	path := acquireVectorPath()
	path.MoveTo(float32(x0*pxPerPt+dx), float32(y0*pxPerPt+dy))
	path.LineTo(float32(x1*pxPerPt+dx), float32(y1*pxPerPt+dy))

	var stroke vector.StrokeOptions
	stroke.Width = width
	stroke.LineCap = vector.LineCapSquare

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/colorChannelMax, float32(cg)/colorChannelMax, float32(cb)/colorChannelMax, float32(ca)/colorChannelMax)

	vector.StrokePath(dst, path, &stroke, &options)
	releaseVectorPath(path)
}

// Corner-shape paths (DisplayOpPath) replay through the vector painter. The
// engine emits the shaped outline of the box as a command list: Stroke false is
// the background fill, painted with the op colour, and Stroke true is the
// border, centred on that same outline with Width. Centring matches how
// OpStrokeRect already treats the border box edge, so the two border kinds land
// the same ink.
//
// A command point is a canvas point, but the engine builds the list before an
// absolute, flex, or grid placement move and then shifts only the op's X and Y.
// The outline always spans the border box, so the painter rebases the list: it
// translates the list's minimum point onto (X, Y) before the op matrix, the
// canvas scale, and the scroll offset. The rerun is deterministic, and it puts
// the ink where the op box says it belongs in every placement case.
//
// The effective opacity rides on the colour scale, so an opacity tick needs no
// rebuild. The engine has no bitmap painter for this kind, so this vector path
// is the only consumer.

// drawPath fills or strokes one corner-shape path. An op with no command
// list, or no command point, paints nothing.
func drawPath(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	corner, ok := op.CornerPathValue()
	if !ok || corner == nil || len(corner.Cmds) == 0 {
		return
	}

	local := acquireVectorPath()
	populateCornerPath(local, corner.Cmds)

	path := acquireVectorPath()

	var addOpts vector.AddPathOptions
	addOpts.GeoM = pathGeoM(op, corner, dx, dy)
	path.AddPath(local, &addOpts)

	releaseVectorPath(local)

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/colorChannelMax, float32(cg)/colorChannelMax, float32(cb)/colorChannelMax, float32(ca)/colorChannelMax)

	if corner.Stroke {
		// Round joins keep a sharp bevel or notch corner inside the bounds; a
		// miter would spike past them. The engine closes every outline, so line
		// caps never show.
		var stroke vector.StrokeOptions
		stroke.Width = pathStrokeWidth(op)
		stroke.LineJoin = vector.LineJoinRound
		vector.StrokePath(dst, path, &stroke, &options)
	} else {
		vector.FillPath(dst, path, nil, &options)
	}

	releaseVectorPath(path)
}

// populateCornerPath converts the engine's command list to a vector path in canvas
// points into path. Every defined verb is covered: move, line, quad, cubic, and close.
// An unknown verb paints nothing for that command rather than guessing a line,
// and a path that never issues a move still draws from the origin, as the
// vector package does.
func populateCornerPath(path *vector.Path, cmds []layout.PathCmd) {
	for i := range cmds {
		cmd := &cmds[i]

		switch cmd.Verb {
		case layout.PathMoveTo:
			x, y := pathPoint(cmd, 0)
			path.MoveTo(x, y)
		case layout.PathLineTo:
			x, y := pathPoint(cmd, 0)
			path.LineTo(x, y)
		case layout.PathQuadTo:
			controlX, controlY := pathPoint(cmd, 0)
			x, y := pathPoint(cmd, 1)
			path.QuadTo(controlX, controlY, x, y)
		case layout.PathCubicTo:
			controlX, controlY := pathPoint(cmd, 0)
			secondX, secondY := pathPoint(cmd, 1)
			x, y := pathPoint(cmd, pathEndPoint)
			path.CubicTo(controlX, controlY, secondX, secondY, x, y)
		case layout.PathClose:
			path.Close()
		}
	}
}

// pathPoint reads one command control point as the vector package's float.
func pathPoint(cmd *layout.PathCmd, index int) (float32, float32) {
	return float32(cmd.Pts[index].X), float32(cmd.Pts[index].Y)
}

// pathGeoM maps an op's command points to device pixels: the rebase onto the op
// origin, the op matrix (the identity when it has none), the canvas scale, and
// the scroll offset. The result is the point mapping the raster path gets from
// xformGeom, applied to the geometry instead of to a baked image.
func pathGeoM(op *layout.DisplayOp, corner *layout.CornerPath, dx, dy float64) ebiten.GeoM {
	rebaseX, rebaseY := pathRebase(op, corner)
	m := pathMatrix(op)

	var geom ebiten.GeoM
	geom.SetElement(0, 0, pxPerPt*m.A)
	geom.SetElement(0, 1, pxPerPt*m.C)
	geom.SetElement(0, geomTransCol, pxPerPt*(m.A*rebaseX+m.C*rebaseY+m.E)+dx)
	geom.SetElement(1, 0, pxPerPt*m.B)
	geom.SetElement(1, 1, pxPerPt*m.D)
	geom.SetElement(1, geomTransCol, pxPerPt*(m.B*rebaseX+m.D*rebaseY+m.F)+dy)

	return geom
}

// pathRebase is the translation from the engine's path space to the op box:
// the list's minimum point moves onto the op origin. An empty list rebases by
// nothing, because it never paints.
func pathRebase(op *layout.DisplayOp, corner *layout.CornerPath) (float64, float64) {
	left, top, _, _, ok := pathRawBox(corner)
	if !ok {
		return 0, 0
	}

	return op.X - left, op.Y - top
}

// pathMatrix is the op's baked matrix, or the identity. op.Transform returns
// the zero matrix when the op carries another payload but no transform, so the
// gate is XformSet, not the accessor's own fallback.
func pathMatrix(op *layout.DisplayOp) layout.Matrix2D {
	if !op.XformSet {
		return layout.IdentityMatrix()
	}

	return op.Transform()
}

// pathStrokeWidth is the border width in device pixels. A matrix scales the
// width by its geometric mean, so a uniform scale grows the stroke with the
// shape; a non-uniform matrix cannot carry an elliptical pen through a vector
// stroke, and the mean is the closest single width.
func pathStrokeWidth(op *layout.DisplayOp) float32 {
	m := pathMatrix(op)
	width := op.Width * pxPerPt

	if det := math.Abs(m.A*m.D - m.B*m.C); det > 0 {
		width *= math.Sqrt(det)
	}

	if width < 1 {
		width = 1
	}

	return float32(width)
}

// pathPointCount is how many control points one command consumes, so bounds
// can walk a path without a second verb switch.
func pathPointCount(verb layout.PathVerb) int {
	switch verb {
	case layout.PathMoveTo, layout.PathLineTo:
		return 1
	case layout.PathQuadTo:
		return pathQuadPoints
	case layout.PathCubicTo:
		return pathCubicPoints
	default:
		return 0
	}
}

// Control and end point counts of the curved and multi-point verbs.
const (
	pathQuadPoints  = 2
	pathCubicPoints = 3
	// pathEndPoint is the end point's slot after a cubic's two control slots.
	pathEndPoint = 2
)

// pathRawBox bounds the command points in the engine's own path space, before
// the rebase and the op matrix.
func pathRawBox(corner *layout.CornerPath) (left, top, right, bottom float64, ok bool) {
	for i := range corner.Cmds {
		cmd := &corner.Cmds[i]

		for index := range pathPointCount(cmd.Verb) {
			x, y := cmd.Pts[index].X, cmd.Pts[index].Y

			if !ok {
				left, top, right, bottom, ok = x, y, x, y, true

				continue
			}

			left, top = math.Min(left, x), math.Min(top, y)
			right, bottom = math.Max(right, x), math.Max(bottom, y)
		}
	}

	return left, top, right, bottom, ok
}

// pathInkBox bounds a corner path in canvas points: the rebased and mapped
// command points, which cover a quad or cubic curve by the convex hull of its
// control points, and the stroke growth on every side. ok is false when the op
// carries no command point, so the caller can skip it as paint-free.
func pathInkBox(op *layout.DisplayOp) (left, top, right, bottom, grow float64, ok bool) {
	corner, present := op.CornerPathValue()
	if !present || corner == nil {
		return 0, 0, 0, 0, 0, false
	}

	rebaseX, rebaseY := pathRebase(op, corner)
	m := pathMatrix(op)

	for i := range corner.Cmds {
		cmd := &corner.Cmds[i]

		for index := range pathPointCount(cmd.Verb) {
			x, y := m.Apply(cmd.Pts[index].X+rebaseX, cmd.Pts[index].Y+rebaseY)

			if !ok {
				left, top, right, bottom, ok = x, y, x, y, true

				continue
			}

			left, top = math.Min(left, x), math.Min(top, y)
			right, bottom = math.Max(right, x), math.Max(bottom, y)
		}
	}

	if !ok {
		return 0, 0, 0, 0, 0, false
	}

	// A fill stops at the command points; a stroke reaches half its width past
	// them, the same growth opBounds gives an OpStrokeRect.
	if corner.Stroke {
		grow = math.Max(op.Width, 0) / halfDiv
	}

	return left, top, right, bottom, grow, true
}
