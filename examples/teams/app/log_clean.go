package app

import "strings"

// cleanLog strips CR/LF so DB errors cannot forge log lines.
func cleanLog(s string) string {
	s = strings.ReplaceAll(s, "\r", " ")
	return strings.ReplaceAll(s, "\n", " ")
}
