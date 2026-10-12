package window

import "time"

// quantile interpolates the q quantile from an ascending slice.
func quantile(s []time.Duration, q float64) time.Duration {
	n := len(s)
	if n == 0 {
		return 0
	}

	at := q * float64(n-1)
	lo := int(at)
	if lo >= n-1 {
		return s[n-1]
	}

	fr := at - float64(lo)

	return s[lo] + time.Duration(fr*float64(s[lo+1]-s[lo]))
}
