package wire

import "strings"

// cleanLog strips CR/LF so store errors cannot forge log lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
