package frame

const (
	// easeMidpoint is the progress where EaseInOutCubic switches halves.
	easeMidpoint = 0.5
	// halfDivisor splits a range into two halves.
	halfDivisor = 2
	// degreesToRadians is the degrees in a half circle.
	degreesToRadians = 180
)

// Linear eases nothing.
func Linear(t float64) float64 { return t }

// EaseOutCubic starts fast and settles.
func EaseOutCubic(t float64) float64 {
	u := 1 - t

	return 1 - u*u*u
}

// EaseInOutCubic accelerates, then settles.
func EaseInOutCubic(t float64) float64 {
	if t < easeMidpoint {
		return 4 * t * t * t
	}

	u := -halfDivisor*t + halfDivisor

	return 1 - u*u*u/2
}

// easeOr returns ease, or Linear when it is nil.
func easeOr(ease Easing) Easing {
	if ease == nil {
		return Linear
	}

	return ease
}
