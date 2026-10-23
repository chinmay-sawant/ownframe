package gesture

// Drag is one claimed pointer capture. The page stores it and does not
// read the fields; Begin, Move, and End are the whole state machine.
type Drag struct {
	active bool
	x, y   float64
}

// Begin records a claimed capture at x, y.
func (d Drag) Begin(x, y float64) Drag {
	return Drag{active: true, x: x, y: y}
}

// Move reports the delta from the last point while the capture is active.
// The position updates even when the pointer did not move. send is false
// when the capture is idle or the delta is zero, so the caller skips the
// move handler. An idle drag is returned unchanged.
func (d Drag) Move(x, y float64) (Drag, float64, float64, bool) {
	if !d.active {
		return d, 0, 0, false
	}

	dx, dy := x-d.x, y-d.y
	d.x, d.y = x, y

	return d, dx, dy, dx != 0 || dy != 0
}

// End reports whether the capture was active and clears it.
func (d Drag) End() (Drag, bool) {
	return Drag{}, d.active
}

// Claimed reports whether a drag-start callback took the gesture.
// An error leaves the capture inactive.
func Claimed(err error, claimed bool) bool {
	return err == nil && claimed
}
