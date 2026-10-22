package page

import (
	"strings"

	"golang.org/x/text/secure/precis"
)

// bidiMarks are U+202A-202E and U+2066-2069. They reorder or isolate
// runs and can disguise untrusted text (Trojan Source style).
var bidiMarks = []string{"\u202a", "\u202b", "\u202c", "\u202d", "\u202e", "\u2066", "\u2067", "\u2068", "\u2069"}

// SanitizeBidi strips bidi controls from untrusted text. Call it on
// chat, notification, and other app-external strings before SetData.
func SanitizeBidi(s string) string {
	for _, mark := range bidiMarks {
		s = strings.ReplaceAll(s, mark, "")
	}
	return s
}

// CheckIdentifier enforces the precis nickname profile on handles,
// room names, and other identifiers rendered from untrusted input.
func CheckIdentifier(s string) error {
	_, err := precis.Nickname.String(s)
	return err
}
