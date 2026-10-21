package page

import "strings"

// cleanReload strips CR/LF so reload errors cannot forge log lines or
// debug state lines. File paths stay but stay on one line.
func cleanReload(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
