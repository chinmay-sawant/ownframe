// Package hit tests a point against layout boxes and builds the id index.
// It does not import package page.
package hit

import "github.com/chinmay-sawant/blinkless/layout"

// Spot is one point query: the innermost box under the point and the
// innermost id-bearing box under it.
type Spot struct {
	Box layout.Box
	OK  bool
	ID  string
}

// contains reports whether the point lies in the box, far edges included.
// The reject form is the rule: a NaN comparison is false, so NaN is not
// thrown out. Do not rewrite this as >=.
func contains(b layout.Box, x, y float64) bool {
	if x < b.X || y < b.Y || x > b.X+b.W || y > b.Y+b.H {
		return false
	}

	return true
}

// Scan walks document-order boxes once and settles both winners, so
// Click, Press, and Hover cannot disagree on overlapping boxes.
func Scan(boxes []layout.Box, x, y float64) Spot {
	var spot Spot
	for _, b := range boxes {
		if !contains(b, x, y) {
			continue
		}

		spot.Box, spot.OK = b, true
		if b.ID != "" {
			spot.ID = b.ID
		}
	}

	return spot
}

// Hit returns the last box that contains x, y.
// Boxes are in document order, so that box is the inner element.
func Hit(boxes []layout.Box, x, y float64) (layout.Box, bool) {
	spot := Scan(boxes, x, y)

	return spot.Box, spot.OK
}

// ScanList returns the box under a point. scrollX and scrollY are the
// window offset in CSS pixels. The window paints a fixed box in viewport
// coordinates, so that box is tested at the point minus the offset and
// wins over the content it covers. Inside one layer the innermost box in
// document order wins.
func ScanList(items []Item, x, y float64, scrollX, scrollY int) Spot {
	vx := x - float64(scrollX)
	vy := y - float64(scrollY)
	content := Spot{}
	pinned := Spot{}

	for _, item := range items {
		px, py := x, y
		spot := &content
		if item.Fixed {
			px, py = vx, vy
			spot = &pinned
		}
		if !contains(item.Box, px, py) {
			continue
		}

		spot.Box = item.Box
		spot.OK = true
		if item.Box.ID != "" {
			spot.ID = item.Box.ID
		}
	}
	if pinned.OK {
		return pinned
	}

	return content
}

// IsFixed reports whether the window pins the box to the viewport.
// An anonymous box is never treated as fixed.
func IsFixed(items []Item, box layout.Box) bool {
	if box.ID == "" {
		return false
	}

	for _, item := range items {
		if item.Fixed && item.Box.ID == box.ID {
			return true
		}
	}

	return false
}
