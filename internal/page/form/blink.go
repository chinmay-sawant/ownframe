package form

import "time"

// BlinkInterval is how long the caret stays shown or hidden.
const BlinkInterval = 530 * time.Millisecond

// ResetBlink shows the caret and restarts the blink clock.
func ResetBlink() (bool, time.Time) {
	return true, time.Time{}
}

// AdvanceBlink starts the clock when at is zero, and toggles once the
// interval has elapsed. The last bool is true when the caller should draw.
func AdvanceBlink(on bool, at, now time.Time) (bool, time.Time, bool) {
	if at.IsZero() {
		return true, now, false
	}

	if now.Sub(at) < BlinkInterval {
		return on, at, false
	}

	return !on, now, true
}
