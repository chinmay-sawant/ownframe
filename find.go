package ownframe

import "github.com/chinmay-sawant/ownframe/internal/frame"

// BoxUnits converts a hit-test box to display-list units: x, y, w, h.
// Display operations carry points; boxes carry CSS pixels.
func BoxUnits(d *Display, b Box) (x, y, w, h float64) {
	return frame.BoxUnits(d, b)
}

// Fill returns the first fill rectangle of the given 0..1 color inside box,
// or nil. Fills returns them all, left to right.
func Fill(d *Display, box Box, color [3]float64) *DisplayOp {
	return frame.Fill(d, box, color)
}

// Fills returns the fill rectangles of the given color inside box, ordered
// left to right, or nil.
func Fills(d *Display, box Box, color [3]float64) []*DisplayOp {
	return frame.Fills(d, box, color)
}

// Text returns the first text operation inside box, or nil. A caller can
// change its Text field to show a new value on the next frame.
func Text(d *Display, box Box) *DisplayOp {
	return frame.Text(d, box)
}
