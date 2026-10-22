package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Time scales and digit grouping for the formatters.
const (
	digitGroup  = 3
	minuteSecs  = 60
	hourSecs    = 3600
	daySecs     = 86400
	truncateMin = 2
)

// formatCount renders an integer with thousands separators.
func formatCount(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= digitGroup {
		return s
	}

	var b strings.Builder

	for i, r := range s {
		if i > 0 && (len(s)-i)%digitGroup == 0 {
			b.WriteByte(',')
		}

		b.WriteRune(r)
	}

	return b.String()
}

// formatAge renders a duration as "3d 4h", "5h 12m", "1m 30s", or "42s".
func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}

	sec := int(d.Seconds())

	switch {
	case sec < minuteSecs:
		return fmt.Sprintf("%ds", sec)
	case sec < hourSecs:
		return fmt.Sprintf("%dm %ds", sec/minuteSecs, sec%minuteSecs)
	case sec < daySecs:
		return fmt.Sprintf("%dh %dm", sec/hourSecs, sec%hourSecs/minuteSecs)
	default:
		return fmt.Sprintf("%dd %dh", sec/daySecs, sec%daySecs/hourSecs)
	}
}

// clock formats a sample time for the header.
func clock(t time.Time) string {
	if t.IsZero() {
		return "no sample yet"
	}

	return t.Format("15:04:05")
}

// truncate shortens s to at most n runes, adding an ellipsis.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}

	if n < truncateMin {
		return string(r[:n])
	}

	return string(r[:n-1]) + "\u2026"
}
