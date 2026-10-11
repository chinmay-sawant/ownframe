package page

import (
	"html"
	"strings"
)

func markedPreedit(value, preedit string, start, end, selStart, selEnd int) string {
	base, composing := []rune(value), []rune(preedit)
	start, end = rangeOf(start, end, len(base))
	if end == start {
		start = clampPos(start, len(base))
	}
	a, b := runeAtByte(preedit, selStart), runeAtByte(preedit, selEnd)
	a, b = rangeOf(a, b, len(composing))
	var out strings.Builder
	out.WriteString(html.EscapeString(string(base[:start])))
	out.WriteString(`<span data-ownframe-composition="1">`)
	out.WriteString(html.EscapeString(string(composing[:a])))
	if a == b {
		out.WriteString(caretSpan())
	} else {
		out.WriteString(selectionSpan(string(composing[a:b])))
	}
	out.WriteString(html.EscapeString(string(composing[b:])))
	out.WriteString(`</span>`)
	out.WriteString(html.EscapeString(string(base[end:])))
	return out.String()
}
