package window

import "strings"

// cleanLog strips CR/LF so reload text cannot forge stderr lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
