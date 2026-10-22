package ui

import "strings"

// timeWidth is the character width of the time column; a raw malformed
// timestamp is cut to it so rows stay aligned.
const timeWidth = 12

// formatTime renders an entry timestamp, falling back to the raw text of a
// malformed one, then to a dashed placeholder.
func formatTime(e Entry) string {
	if e.TimeOK && !e.Time.IsZero() {
		return e.Time.Format("15:04:05.000")
	}

	if e.TimeRaw != "" {
		return truncate(firstLine(e.TimeRaw), timeWidth)
	}

	return "--:--:--"
}

// sevClass maps a severity label to one of the CSS class names.
func sevClass(sev string) string {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case "trace", "trc", "verbose":
		return "trace"
	case "warn", "wrn", "warning":
		return "warn"
	case "error", "err", "severe":
		return "error"
	case "fatal", "crit", "critical", "panic", "alert", "emerg":
		return "fatal"
	case "info", "inf", "information", "notice":
		return "info"
	}

	return "debug"
}

// sevLabel returns the compact uppercase label for a severity.
func sevLabel(sev string) string {
	class := sevClass(sev)
	if class == "debug" {
		return "DBG"
	}

	return strings.ToUpper(class[:3])
}

// firstLine returns s up to the first line break.
func firstLine(s string) string {
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		return s[:i]
	}

	return s
}
