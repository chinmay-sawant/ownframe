package reload

import "strings"

// Clean strips CR/LF so a reload error cannot forge a log line. The text
// stays on one line.
func Clean(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
