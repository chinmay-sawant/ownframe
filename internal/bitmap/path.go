package bitmap

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Corner-shape paths (DisplayOpPath) paint through the same four-sample
// coverage loop as a rectangle, with a point-in-path test in place of the
// rounded-box test. The engine emits one path for the background fill and one
// for the border stroke; both carry the shaped outer outline of the border
// box. The command list is flattened here rather than handed to the vector
// renderer, so the fallback picture stays independent of the replay painter.
//
// The engine builds the command list before an absolute, flex, or grid
// placement move and then shifts only the op's X and Y. The list is therefore
// rebased: its minimum point lands on (X, Y), the same rule the replay painter
// uses. Without the rebase, a path in a placed box would paint at the page
// origin.
//
// The effective opacity rides on the colour, through opColor, so an opacity
// tick needs no rebuild. The op matrix maps the grown local box for the bounds
// and is inverted per sample, so a transformed box paints through its matrix.

// pathCurveSteps is how many line segments one quadratic or cubic command
// flattens to. A corner arc spans a quarter turn at most, so sixteen segments
// keep the chord error far below one pixel at the radii a page uses.
const pathCurveSteps = 16

// pathPad is one pixel of slack around the painted bounds for antialiasing,
// matching the replay bounds rule.
const pathPad = 1

// pathPoint is one flattened path point in the operation's local point space.
type pathPoint struct {
	x, y float64
}

// pathContour is one flattened contour. closed records a PathClose, so a
// stroke does not invent a closing edge the engine never asked for.
type pathContour struct {
	points []pathPoint
	closed bool
}

// paintPath fills or strokes one corner-shape path. An op with no command
// list, or no command point, paints nothing.
func paintPath(img *image.NRGBA, op *layout.DisplayOp) {
	corner, ok := op.CornerPathValue()
	if !ok || corner == nil || len(corner.Cmds) == 0 {
		return
	}

	contours := flattenCornerPath(op, corner)
	if len(contours) == 0 {
		return
	}

	half := 0.0
	if corner.Stroke {
		half = pathStrokeHalf(op)
	}

	bounds := pathPaintBounds(op, contours, half)
	hit := func(x, y float64) bool {
		if corner.Stroke {
			return pathStroked(contours, x, y, half)
		}

		return pathFilled(contours, x, y)
	}

	paintCoverage(img, op, bounds, hit)
}

// flattenCornerPath converts the engine's command list to contours in the
// operation's local point space, rebased onto the op origin. Every defined verb
// is covered. An unknown verb is skipped for that command rather than guessed
// as a line, and a contour that never issues a move starts at the rebased path
// origin, as the vector package starts it at the path-space origin.
func flattenCornerPath(op *layout.DisplayOp, corner *layout.CornerPath) []pathContour {
	originX, originY, ok := pathMinPoint(corner.Cmds)
	if !ok {
		return nil
	}

	shiftX, shiftY := op.X-originX, op.Y-originY
	rebased := func(p layout.DisplayPoint) pathPoint {
		return pathPoint{x: p.X + shiftX, y: p.Y + shiftY}
	}

	var (
		contours []pathContour
		current  = -1
		last     = pathPoint{x: shiftX, y: shiftY}
	)

	add := func(p pathPoint) {
		if current < 0 {
			contours = append(contours, pathContour{})
			current = len(contours) - 1
		}

		contours[current].points = append(contours[current].points, p)
		last = p
	}

	for i := range corner.Cmds {
		cmd := &corner.Cmds[i]

		switch cmd.Verb {
		case layout.PathMoveTo:
			contours = append(contours, pathContour{points: []pathPoint{rebased(cmd.Pts[0])}})
			current = len(contours) - 1
			last = contours[current].points[0]
		case layout.PathLineTo:
			add(rebased(cmd.Pts[0]))
		case layout.PathQuadTo:
			control, end := rebased(cmd.Pts[0]), rebased(cmd.Pts[1])
			from := last
			for step := 1; step <= pathCurveSteps; step++ {
				add(quadPoint(from, control, end, float64(step)/pathCurveSteps))
			}
		case layout.PathCubicTo:
			first, second, end := rebased(cmd.Pts[0]), rebased(cmd.Pts[1]), rebased(cmd.Pts[pathCubicEnd])
			from := last
			for step := 1; step <= pathCurveSteps; step++ {
				add(cubicPoint(from, first, second, end, float64(step)/pathCurveSteps))
			}
		case layout.PathClose:
			if current >= 0 {
				contours[current].closed = true
				last = contours[current].points[0]
			}
		}
	}

	return contours
}

// pathCubicEnd is the end point's slot after a cubic's two control points.
const pathCubicEnd = 2

// Control-point counts of the path verbs.
const (
	pathPointPer    = 1
	pathQuadPoints  = 2
	pathCubicPoints = 3
	// pathStrokePoints is the fewest flattened points a contour needs before
	// it has an edge to stroke.
	pathStrokePoints = 2
)

// pathMinPoint bounds the command points in the engine's path space and
// reports the minimum, which is the rebase origin. ok is false when the list
// carries no point.
func pathMinPoint(cmds []layout.PathCmd) (x, y float64, ok bool) {
	for i := range cmds {
		cmd := &cmds[i]

		for index := range pathCmdPoints(cmd.Verb) {
			px, py := cmd.Pts[index].X, cmd.Pts[index].Y

			if !ok {
				x, y, ok = px, py, true

				continue
			}

			x, y = math.Min(x, px), math.Min(y, py)
		}
	}

	return x, y, ok
}

// pathCmdPoints is how many control points one command consumes.
func pathCmdPoints(verb layout.PathVerb) int {
	switch verb {
	case layout.PathMoveTo, layout.PathLineTo:
		return pathPointPer
	case layout.PathQuadTo:
		return pathQuadPoints
	case layout.PathCubicTo:
		return pathCubicPoints
	}

	return 0
}

// pathFilled reports whether a point is inside the shaped outline, using the
// nonzero winding rule the canvas fill applies. Contours are implicitly closed
// for a fill, as canvas closes every subpath it fills.
func pathFilled(contours []pathContour, x, y float64) bool {
	winding := 0

	for i := range contours {
		points := contours[i].points
		for j := range points {
			from, to := points[j], points[(j+1)%len(points)]

			switch {
			case from.y <= y && to.y > y && pathSide(from, to, x, y) > 0:
				winding++
			case from.y > y && to.y <= y && pathSide(from, to, x, y) < 0:
				winding--
			}
		}
	}

	return winding != 0
}

// pathStroked reports whether a point is within half the stroke width of the
// flattened outline. Distance to the nearest segment gives round joins and
// caps for free, matching the replay painter's round line join. An unclosed
// contour has no closing edge.
func pathStroked(contours []pathContour, x, y, half float64) bool {
	for i := range contours {
		points := contours[i].points
		if len(points) < pathStrokePoints {
			continue
		}

		for j := range points {
			next := j + 1
			if next == len(points) {
				if !contours[i].closed {
					break
				}

				next = 0
			}

			if pathSegmentDistance(x, y, points[j], points[next]) <= half {
				return true
			}
		}
	}

	return false
}

// pathSide is the signed area of the triangle the point makes with one edge,
// the winding test's left-of-edge predicate.
func pathSide(from, to pathPoint, x, y float64) float64 {
	return (to.x-from.x)*(y-from.y) - (x-from.x)*(to.y-from.y)
}

// pathSegmentDistance is the distance from a point to one straight edge.
func pathSegmentDistance(x, y float64, from, to pathPoint) float64 {
	dx, dy := to.x-from.x, to.y-from.y
	if dx == 0 && dy == 0 {
		return math.Hypot(x-from.x, y-from.y)
	}

	t := ((x-from.x)*dx + (y-from.y)*dy) / (dx*dx + dy*dy)
	t = min(max(t, 0), 1)

	return math.Hypot(x-(from.x+t*dx), y-(from.y+t*dy))
}

// quadPoint evaluates one quadratic Bezier point by de Casteljau, so the
// weights stay in the interpolation rather than as magic factors.
func quadPoint(from, control, to pathPoint, t float64) pathPoint {
	return lerpPoint(lerpPoint(from, control, t), lerpPoint(control, to, t), t)
}

// cubicPoint evaluates one cubic Bezier point by de Casteljau.
func cubicPoint(from, first, second, to pathPoint, t float64) pathPoint {
	a := lerpPoint(from, first, t)
	b := lerpPoint(first, second, t)
	c := lerpPoint(second, to, t)

	return lerpPoint(lerpPoint(a, b, t), lerpPoint(b, c, t), t)
}

// lerpPoint interpolates two points at t.
func lerpPoint(from, to pathPoint, t float64) pathPoint {
	return pathPoint{x: from.x + (to.x-from.x)*t, y: from.y + (to.y-from.y)*t}
}

// pathStrokeHalf is half the border width in the op's local point space. The
// op width and the command points share that space, so the half width is used
// as-is: the op matrix then maps the round pen into an elliptical one, which is
// what a transformed border should look like. No display scale enters here,
// because a sample point is already converted back to points.
//
// The width floors at one device pixel, the same floor the rectangle painter
// gives a hairline border, so a sub-pixel border stays visible.
func pathStrokeHalf(op *layout.DisplayOp) float64 {
	return math.Max(op.Width, 1/pxPerPt) / halfDivisor
}

// pathPaintBounds bounds the shaped ink in device pixels, grown by the stroke
// half-width and one pixel of slack. The local box is grown before the op
// matrix, so a non-uniform matrix cannot clip the ellipse its pen turns into. A
// linear map keeps the image of a box inside the mapped corners, so bounding
// those four corners covers every contour point. An op with no point gets an
// empty box.
func pathPaintBounds(op *layout.DisplayOp, contours []pathContour, half float64) image.Rectangle {
	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for i := range contours {
		for _, p := range contours[i].points {
			loX, loY = math.Min(loX, p.x), math.Min(loY, p.y)
			hiX, hiY = math.Max(hiX, p.x), math.Max(hiY, p.y)
		}
	}

	if loX > hiX {
		return image.Rectangle{}
	}

	loX, loY = loX-half, loY-half
	hiX, hiY = hiX+half, hiY+half

	mappedX0, mappedY0 := math.Inf(1), math.Inf(1)
	mappedX1, mappedY1 := math.Inf(-1), math.Inf(-1)

	for _, corner := range [4][2]float64{{loX, loY}, {hiX, loY}, {hiX, hiY}, {loX, hiY}} {
		x, y := transform(op, corner[0], corner[1])
		mappedX0, mappedY0 = math.Min(mappedX0, x), math.Min(mappedY0, y)
		mappedX1, mappedY1 = math.Max(mappedX1, x), math.Max(mappedY1, y)
	}

	return image.Rect(
		int(math.Floor(mappedX0*pxPerPt))-pathPad, int(math.Floor(mappedY0*pxPerPt))-pathPad,
		int(math.Ceil(mappedX1*pxPerPt))+pathPad, int(math.Ceil(mappedY1*pxPerPt))+pathPad,
	)
}
