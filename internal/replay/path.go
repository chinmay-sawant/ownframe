package replay

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

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

	local := cornerPathAt(corner.Cmds)
	var path vector.Path
	path.AddPath(&local, &vector.AddPathOptions{GeoM: pathGeoM(op, corner, dx, dy)})

	options := &vector.DrawPathOptions{AntiAlias: true}
	options.ColorScale.ScaleWithColor(rgba(op))

	if corner.Stroke {
		// Round joins keep a sharp bevel or notch corner inside the bounds; a
		// miter would spike past them. The engine closes every outline, so line
		// caps never show.
		stroke := &vector.StrokeOptions{Width: pathStrokeWidth(op), LineJoin: vector.LineJoinRound}
		vector.StrokePath(dst, &path, stroke, options)

		return
	}

	vector.FillPath(dst, &path, nil, options)
}

// cornerPathAt converts the engine's command list to a vector path in canvas
// points. Every defined verb is covered: move, line, quad, cubic, and close.
// An unknown verb paints nothing for that command rather than guessing a line,
// and a path that never issues a move still draws from the origin, as the
// vector package does.
func cornerPathAt(cmds []layout.PathCmd) vector.Path {
	var path vector.Path

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

	return path
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
