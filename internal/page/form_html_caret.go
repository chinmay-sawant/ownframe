package page

import (
	"html"
	"strings"
)

// fieldSpanState rewrites one text-like control into a span. It paints the
// range as a selection span and the whole-value select as the field
// background. The page paints the caret after layout.
func fieldSpanState(raw string, ctrl Control, st caretState, focused bool) string {
	kind := strings.ToLower(strings.TrimSpace(ctrl.Type))
	tag := strings.ToLower(strings.TrimSpace(ctrl.Tag))
	text := shownText(kind, ctrl.Value)
	ph := ""
	if text == "" && kind != "file" {
		ph = attrValue(raw, "placeholder")
	}
	if st.preedit != "" {
		ph = ""
	}

	start, end := rangeOf(st.start, st.end, runeLen(ctrl.Value))
	whole := focused && (st.all || (start != end && start == 0 && end == runeLen(ctrl.Value)))

	extras := []string{`data-ownframe-field="` + tag + `"`}
	if focused {
		extras = append(extras, focusAttr)
	}
	if whole {
		extras = append(extras, selectedAttr)
	}

	inner := ""
	switch {
	case ph != "":
		extras = append(extras, placeholderAttr)
		inner = html.EscapeString(ph)
	case st.preedit != "":
		inner = markedPreedit(text, shownText(kind, st.preedit), st.start,
			st.end, st.preeditStart, st.preeditEnd)
	case whole:
		inner = selectionSpan(text)
	default:
		inner = markedText(text, focused, start, end)
	}

	return openTagDrop("span", raw, extras, spanDrop) + inner + "</span>"
}
