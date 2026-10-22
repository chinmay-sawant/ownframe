package ui

import (
	"strconv"
	"strings"
)

// lineCount counts the lines in s, one for an empty string.
func lineCount(s string) int {
	if s == "" {
		return 1
	}

	n := 1
	for _, r := range s {
		if r == '\n' {
			n++
		}
	}

	return n
}

// splitLines splits a bounded message into display lines.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")

	return strings.Split(s, "\n")
}

// truncate cuts s to at most n runes, appending an ellipsis when cut.
func truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= n {
		return s
	}

	return string(runes[:n-1]) + "…"
}

// bytesText formats a byte count for the detail header.
func bytesText(n int64) string {
	switch {
	case n >= 1<<20:
		return strconv.FormatFloat(float64(n)/(1<<uiMiBShift), 'f', 1, 64) + "M"
	case n >= 1<<10:
		return strconv.FormatFloat(float64(n)/(1<<uiKiBShift), 'f', 1, 64) + "K"
	}

	return strconv.FormatInt(n, 10) + "B"
}

// textBudget estimates how many runes fit in the message column of a page
// width. The estimate is deliberately conservative: a wrapped row would
// break the fixed-height row contract the window relies on.
func textBudget(width int) int {
	const fixed = 500 // sidebar, padding, time, source, severity, tags

	if width < fixed+uiMinWidth {
		width = fixed + uiMinWidth
	}

	budget := (width - fixed) / uiColChar
	if budget > uiMaxBudget {
		budget = uiMaxBudget
	}

	return budget
}
