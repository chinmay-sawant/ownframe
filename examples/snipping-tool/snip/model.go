// Package snip is the pure selection/crop/annotate model for the
// snipping-tool example. It keeps no window, clipboard, or network
// state, so its tests stay headless.
package snip

// Point is a pixel in absolute coordinates.
type Point struct{ X, Y int }

// Rect is a normalized rectangle: W and H are never negative.
type Rect struct{ X, Y, W, H int }

// Normalize returns the bounding box of drag endpoints a and b.
func Normalize(a, b Point) Rect {
	x, w := a.X, b.X-a.X
	if w < 0 {
		x, w = b.X, -w
	}
	y, h := a.Y, b.Y-a.Y
	if h < 0 {
		y, h = b.Y, -h
	}
	return Rect{X: x, Y: y, W: w, H: h}
}

// Clamp intersects r with bounds. A non-overlapping r keeps its
// clamped origin with zero size.
func Clamp(r, bounds Rect) Rect {
	x1 := max(r.X, bounds.X)
	y1 := max(r.Y, bounds.Y)
	x2 := min(r.X+r.W, bounds.X+bounds.W)
	y2 := min(r.Y+r.H, bounds.Y+bounds.H)
	return Rect{X: x1, Y: y1, W: max(0, x2-x1), H: max(0, y2-y1)}
}

// Selection is an in-progress drag from Start to End.
type Selection struct {
	Start, End Point
	Active     bool
}

// Rect normalizes the drag endpoints.
func (s Selection) Rect() Rect { return Normalize(s.Start, s.End) }
