package clipboard

import "strings"

// cleanLog strips CR/LF so clipboard panic text cannot forge log lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
