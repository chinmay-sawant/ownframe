package player

import (
	"fmt"
	"strconv"
	"strings"
)

// formatMillis turns iTunes trackTimeMillis into m:ss.
func formatMillis(ms int) string {
	if ms <= 0 {
		return "0:00"
	}

	return formatSeconds(ms / millisPerSec)
}

// formatSeconds turns a count of seconds into m:ss.
func formatSeconds(sec int) string {
	if sec < 0 {
		sec = 0
	}

	return fmt.Sprintf("%d:%02d", sec/secsPerMin, sec%secsPerMin)
}

// lengthSeconds parses an m:ss length back into seconds.
func lengthSeconds(length string) int {
	parts := strings.SplitN(length, ":", timeParts)
	if len(parts) != timeParts {
		return 0
	}

	m, _ := strconv.Atoi(parts[0])
	s, _ := strconv.Atoi(parts[1])

	return m*60 + s
}

// yearOf keeps the first four characters of an iTunes release date.
func yearOf(date string) string {
	if len(date) < dateMinLen {
		return date
	}

	return date[:4]
}

// clip shortens s to n runes with an ellipsis, so a long iTunes title
// cannot run past its card.
func clip(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n-1]) + "…"
}
