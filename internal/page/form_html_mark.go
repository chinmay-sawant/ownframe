package page

import "html"

// selectionSpan wraps selected runes in the library selection element, so
// the highlight paints inside the field and author field backgrounds cannot
// cover it.
func selectionSpan(text string) string {
	return `<span data-ownframe-selection="1">` + html.EscapeString(text) + `</span>`
}

// markedText leaves caret placement to the display-list overlay so adding a
// caret never splits or reflows the field's text run.
func markedText(text string, focused bool, start, end int) string {
	if !focused {
		return html.EscapeString(text)
	}

	runes := []rune(text)
	if end > start {
		return html.EscapeString(string(runes[:start])) +
			selectionSpan(string(runes[start:end])) +
			html.EscapeString(string(runes[end:]))
	}

	return html.EscapeString(text)
}
