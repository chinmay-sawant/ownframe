package frame

import (
	"math"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// centerDivisor halves a size to reach its center.
const centerDivisor = 2

// inside reports whether op's center lies in the box, with a half-point of
// slack for rounded edges.
func inside(op *page.DisplayOp, x, y, w, h float64) bool {
	cx, cy := op.X+op.W/centerDivisor, op.Y+op.H/centerDivisor

	return cx >= x-0.5 && cx <= x+w+0.5 && cy >= y-0.5 && cy <= y+h+0.5
}

// sameColor compares an op's paint to a 0..1 color within one 8-bit step.
func sameColor(op *page.DisplayOp, color [3]float64) bool {
	const tolerance = 0.02

	return op.Alpha > 0.5 &&
		math.Abs(op.R-color[0]) < tolerance &&
		math.Abs(op.G-color[1]) < tolerance &&
		math.Abs(op.B-color[2]) < tolerance
}
