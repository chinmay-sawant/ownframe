// Package frame finds and changes display-list operations between frames, so
// a page can animate without another Redraw. The window draws every frame;
// changing an operation's paint fields changes what the next frame shows.
package frame

import (
	"slices"

	"github.com/chinmay-sawant/ownframe"
)

// BoxUnits converts a hit-test box to display-list units: x, y, w, h.
// Display operations carry points; boxes carry CSS pixels.
func BoxUnits(d *ownframe.Display, b ownframe.Box) (x, y, w, h float64) {
	p := d.PointsPerPixel

	return b.X * p, b.Y * p, b.W * p, b.H * p
}

// Fill returns the first fill rectangle of the given 0..1 color inside box,
// or nil. Fills returns them all, left to right.
func Fill(d *ownframe.Display, box ownframe.Box, color [3]float64) *ownframe.DisplayOp {
	fills := Fills(d, box, color)
	if len(fills) == 0 {
		return nil
	}

	return fills[0]
}

// Fills returns the fill rectangles of the given color inside box, ordered
// left to right, or nil.
func Fills(d *ownframe.Display, box ownframe.Box, color [3]float64) []*ownframe.DisplayOp {
	if d == nil {
		return nil
	}

	x, y, w, h := BoxUnits(d, box)
	out := make([]*ownframe.DisplayOp, 0, 4)

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect || !inside(op, x, y, w, h) {
			continue
		}

		if !sameColor(op, color) {
			continue
		}

		out = append(out, op)
	}

	slices.SortFunc(out, func(a, b *ownframe.DisplayOp) int {
		switch {
		case a.X < b.X:
			return -1
		case a.X > b.X:
			return 1
		default:
			return 0
		}
	})

	return out
}
