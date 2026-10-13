package frame

// Linear eases nothing.
func Linear(t float64) float64 { return t }

// EaseOutCubic starts fast and settles.
func EaseOutCubic(t float64) float64 {
	u := 1 - t

	return 1 - u*u*u
}

// EaseInOutCubic accelerates, then settles.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}

	u := -2*t + 2

	return 1 - u*u*u/2
}

// easeOr returns ease, or Linear when it is nil.
func easeOr(ease Easing) Easing {
	if ease == nil {
		return Linear
	}

	return ease
}
