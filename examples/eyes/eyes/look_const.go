package eyes

import "time"

// Look bounds and pointer gains in CSS pixels.
const (
	lookXRange  = 48
	lookYRange  = 30
	saccadeX    = 44
	saccadeY    = 26
	squintShift = 6
	pressGain   = 0.12
	dragGain    = 0.4
	swipeGain   = 0.2
	tiltGain    = 96 // maps about half a g to the full look range
	keyLookX    = 44
	keyLookY    = 26
	unitSpan    = 2 // maps [0,1) to [-1,1)
)

// halfDivisor centers a span or halves a cosine swing.
const halfDivisor = 2

// swipeCooldown delays the next auto saccade after a swipe.
const swipeCooldown = 2500 * time.Millisecond
