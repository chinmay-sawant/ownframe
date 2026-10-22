package parser

import "strings"

// fracLayout returns the layout piece for a captured fraction, capped at
// nanosecond precision.
func fracLayout(frac string) string {
	if len(frac) < fracMinLen {
		return ""
	}

	n := len(frac) - 1
	if n > fracMaxPrec {
		n = fracMaxPrec
	}

	return "." + strings.Repeat("0", n)
}

func fracText(frac string) string {
	if len(frac) > fracMaxLen {
		return frac[:fracMaxLen]
	}

	return frac
}

func fracNanos(frac string) string {
	frac = strings.TrimPrefix(frac, ".")
	for len(frac) < fracPadWidth {
		frac += "0"
	}

	return frac[:fracPadWidth]
}

func zoneLayout(zone string) string {
	switch {
	case zone == "Z", strings.Contains(zone, ":"):
		return "Z07:00"
	case zone != "":
		return "Z0700"
	}

	return ""
}

func parseInt(s string) int64 {
	var n int64

	for _, r := range s {
		if r < '0' || r > '9' {
			return n
		}

		n = n*fracBase + int64(r-'0')
	}

	return n
}
