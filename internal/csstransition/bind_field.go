package csstransition

import (
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// adjust returns the captured paint with every field the operation changed
// since the capture set to the operation's current value. The mask and the
// matrix are captured values, not ranges: a transition keeps the fields its
// property owns, and the matrix pair is what Blend decomposes.
func adjust(captured, initial, final, current cssmotion.Paint) cssmotion.Paint {
	return cssmotion.Paint{
		X:         field(captured.X, initial.X, final.X, current.X),
		Y:         field(captured.Y, initial.Y, final.Y, current.Y),
		W:         field(captured.W, initial.W, final.W, current.W),
		H:         field(captured.H, initial.H, final.H, current.H),
		R:         field(captured.R, initial.R, final.R, current.R),
		G:         field(captured.G, initial.G, final.G, current.G),
		B:         field(captured.B, initial.B, final.B, current.B),
		Alpha:     field(captured.Alpha, initial.Alpha, final.Alpha, current.Alpha),
		Opacity:   field(captured.Opacity, initial.Opacity, final.Opacity, current.Opacity),
		Matrix:    captured.Matrix,
		HasMatrix: captured.HasMatrix,
		Mask:      captured.Mask,
	}
}

// field returns the captured value when the operation's current value still
// lies inside the range the capture spans, and the current value otherwise.
func field(captured, initial, final, current float64) float64 {
	lo, hi := initial, final
	if lo > hi {
		lo, hi = hi, lo
	}

	if current < lo-paintSlack || current > hi+paintSlack {
		return current
	}

	return captured
}
