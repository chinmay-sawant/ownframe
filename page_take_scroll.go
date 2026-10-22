package ownframe

// Scroll is one queued window scroll request.
type Scroll struct {
	X, Y     int
	Absolute bool
}
