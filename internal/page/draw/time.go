package draw

import "time"

// Clock is now when track is on, zero otherwise.
func Clock(track bool) time.Time {
	if !track {
		return time.Time{}
	}

	return time.Now()
}

// Note records since start into dst only when track is on.
func Note(track bool, dst *time.Duration, start time.Time) {
	if track {
		*dst = time.Since(start)
	}
}
