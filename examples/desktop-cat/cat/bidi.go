package cat

import "strings"

// bidiMarks lists U+202A-202E and U+2066-2069. They reorder or isolate
// runs and can disguise bubble text (Trojan Source style).
const bidiMarks = "\u202a\u202b\u202c\u202d\u202e\u2066\u2067\u2068\u2069"

// sanitizeBidi strips bidi controls from agent-supplied text before
// the bubble renders it.
func sanitizeBidi(s string) string {
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(bidiMarks, r) {
			return -1
		}
		return r
	}, s)
}
