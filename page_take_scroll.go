package ownframe

// ScrollRequest is one queued window scroll request.
type ScrollRequest struct {
	X, Y     int
	Absolute bool
}
