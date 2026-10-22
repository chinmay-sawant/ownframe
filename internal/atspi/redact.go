package atspi

import "strings"

// Summarize describes an object for logs. It never includes Name,
// Description, or Value, so default logs do not carry accessible text.
func Summarize(obj Object) string {
	var b strings.Builder
	b.WriteString("atspi ")
	b.WriteString(string(obj.Role))
	b.WriteByte(' ')
	b.WriteString(obj.Path)
	if obj.Password {
		b.WriteString(" password")
	}
	return b.String()
}
