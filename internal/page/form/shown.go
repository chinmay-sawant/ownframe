package form

import (
	"strings"
	"unicode/utf8"
)

// ShownText is the painted field text. A password is bullets; an empty
// file field is "No file".
func ShownText(kind, value string) string {
	if kind == "password" {
		return strings.Repeat("•", utf8.RuneCountInString(value))
	}
	if kind == "file" && value == "" {
		return "No file"
	}

	return value
}
