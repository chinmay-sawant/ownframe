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

// pathSeg is one precomputed flattened segment with its bounding box and
// reciprocal squared length so stroke distance and winding tests skip
// non-overlapping edges without calling math.Hypot.
type pathSeg struct {
	x0, y0, x1, y1 float64
	dx, dy         float64
	invLen2        float64
	minX, maxX     float64
	minY, maxY     float64
}

func buildPathSegs(contours []pathContour, stroke bool) []pathSeg {
	total := 0
	for i := range contours {
		total += len(contours[i].points)
	}
	segs := make([]pathSeg, 0, total)
	for i := range contours {
		points := contours[i].points
		if stroke && len(points) < pathStrokePoints {
			continue
		}
		for j := range points {
			next := j + 1
			if next == len(points) {
				if stroke && !contours[i].closed {
					break
				}
				next = 0
			}
			from, to := points[j], points[next]
			dx, dy := to.x-from.x, to.y-from.y
			len2 := dx*dx + dy*dy
			invLen2 := 0.0
			if len2 > 0 {
				invLen2 = 1.0 / len2
			}
			segs = append(segs, pathSeg{
				x0: from.x, y0: from.y, x1: to.x, y1: to.y,
				dx: dx, dy: dy, invLen2: invLen2,
				minX: min(from.x, to.x), maxX: max(from.x, to.x),
				minY: min(from.y, to.y), maxY: max(from.y, to.y),
			})
		}
	}

	return segs
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

	bounds := pathPaintBounds(op, contours, half).Intersect(img.Bounds())
	if bounds.Empty() {
		return
	}
	segs := buildPathSegs(contours, corner.Stroke)
	if len(segs) == 0 {
		return
	}

	if !op.XformSet {
		paintAxisPath(img, op, bounds, segs, corner.Stroke, half)
		return
	}

	half2 := half * half
	hit := func(x, y float64) bool {
		if corner.Stroke {
			return pathSegsStroked(segs, x, y, half, half2)
		}

		return pathSegsFilled(segs, x, y)
	}

	paintCoverage(img, op, bounds, hit)
}

// paintAxisPath accelerates axis-aligned paths by pre-filtering segments per
// pixel row before evaluating the 4 subpixel samples, keeping exact parity
// with paintCoverage while avoiding O(pixels * all_segments) scans.
func paintAxisPath(img *image.NRGBA, op *layout.DisplayOp, bounds image.Rectangle,
	segs []pathSeg, stroke bool, half float64,
) {
	c := opColor(op)
	if c.A == 0 {
		return
	}
	const (
		subLo = 0.25
		subHi = 0.75
	)
	half2 := half * half
	active := make([]pathSeg, 0, len(segs))

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		py0 := (float64(y) + subLo) / pxPerPt
		py1 := (float64(y) + subHi) / pxPerPt
		rowMinY, rowMaxY := py0-half, py1+half

		active = active[:0]
		rowMinX, rowMaxX := math.Inf(1), math.Inf(-1)
		for i := range segs {
			s := &segs[i]
			if s.maxY < rowMinY || s.minY > rowMaxY {
				continue
			}
			active = append(active, *s)
			rowMinX = min(rowMinX, s.minX)
			rowMaxX = max(rowMaxX, s.maxX)
		}
		if len(active) == 0 {
			continue
		}

		xStart := bounds.Min.X
		xEnd := bounds.Max.X
		if stroke {
			xStart = max(xStart, int(math.Floor((rowMinX-half)*pxPerPt)))
			xEnd = min(xEnd, int(math.Ceil((rowMaxX+half)*pxPerPt)))
		} else {
			xStart = max(xStart, int(math.Floor(rowMinX*pxPerPt)))
			xEnd = min(xEnd, int(math.Ceil(rowMaxX*pxPerPt)))
		}

		for x := xStart; x < xEnd; x++ {
			px0 := (float64(x) + subLo) / pxPerPt
			px1 := (float64(x) + subHi) / pxPerPt
			count := 0
			if stroke {
				if pathSegsStroked(active, px0, py0, half, half2) {
					count++
				}
				if pathSegsStroked(active, px1, py0, half, half2) {
					count++
				}
				if pathSegsStroked(active, px0, py1, half, half2) {
					count++
				}
				if pathSegsStroked(active, px1, py1, half, half2) {
					count++
				}
			} else {
				if pathSegsFilled(active, px0, py0) {
					count++
				}
				if pathSegsFilled(active, px1, py0) {
					count++
				}
				if pathSegsFilled(active, px0, py1) {
					count++
				}
				if pathSegsFilled(active, px1, py1) {
					count++
				}
			}
			if count == 0 {
				continue
			}
			src := c
			if count < shapeSamples {
				src.A = uint8(float64(src.A) * float64(count) / shapeSamples)
			}
			overPixel(img, x, y, src)
		}
	}
}

func pathSegsFilled(segs []pathSeg, x, y float64) bool {
	winding := 0
	for i := range segs {
		s := &segs[i]
		if s.y0 <= y {
			if s.y1 > y && (s.dx*(y-s.y0)-(x-s.x0)*s.dy) > 0 {
				winding++
			}
		} else if s.y1 <= y && (s.dx*(y-s.y0)-(x-s.x0)*s.dy) < 0 {
			winding--
		}
	}

	return winding != 0
}

func pathSegsStroked(segs []pathSeg, x, y, half, half2 float64) bool {
	for i := range segs {
		s := &segs[i]
		if x < s.minX-half || x > s.maxX+half || y < s.minY-half || y > s.maxY+half {
			continue
		}
		if s.invLen2 == 0 {
			rx, ry := x-s.x0, y-s.y0
			if rx*rx+ry*ry <= half2 {
				return true
			}
			continue
		}
		t := ((x-s.x0)*s.dx + (y-s.y0)*s.dy) * s.invLen2
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}
		rx := x - (s.x0 + t*s.dx)
		ry := y - (s.y0 + t*s.dy)
		if rx*rx+ry*ry <= half2 {
			return true
		}
	}

	return false
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
