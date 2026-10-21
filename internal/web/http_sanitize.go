package web

import "strings"

// sanitizeMaxLen caps sanitized strings so logs stay bounded.
const sanitizeMaxLen = 500

// sanitizeLine strips CR/LF so error text cannot forge log lines or
// debug JSON lines. It truncates to keep logs bounded.
func sanitizeLine(s string) string {
	s = strings.ReplaceAll(s, "\r", "_")
	s = strings.ReplaceAll(s, "\n", "_")
	if len(s) > sanitizeMaxLen {
		s = s[:sanitizeMaxLen]
	}
	return s
}

// sanitizeReload applies sanitizeLine to a reload error exposed
// through /debug/state, which may otherwise carry file paths.
func sanitizeReload(s string) string {
	return sanitizeLine(s)
}
