package window

import "github.com/chinmay-sawant/blinkless/layout"

// devOpAt returns the topmost painted operation whose box holds the point
// and whose centre sits in the picked box, or -1 when none does. The point
// rule keeps a click on a title, an image, or a row's background on the
// operation the pointer actually touched.
func devOpAt(display *layout.Display, box layout.Box, x, y float64) int {
	if display == nil {
		return -1
	}

	found := -1

	for _, i := range display.Order {
		if i < 0 || i >= len(display.Ops) {
			continue
		}

		op := &display.Ops[i]
		if _, ok := devOpInk(op.Kind); !ok {
			continue
		}

		r := devOpRect(op, display.PointsPerPixel)
		if x < r.X || y < r.Y || x > r.X+r.W || y > r.Y+r.H {
			continue
		}

		cx, cy := r.X+r.W/2, r.Y+r.H/2
		if cx >= box.X-0.5 && cx <= box.X+box.W+0.5 &&
			cy >= box.Y-0.5 && cy <= box.Y+box.H+0.5 {
			found = i
		}
	}

	return found
}
