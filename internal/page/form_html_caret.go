package page

import (
	"html"
	"strings"
)

// fieldSpanState rewrites one text-like control into a span. It paints the
// caret at st.caret, or the range as a selection span, and the whole-value
// select as the field background. A negative caret means the end of the
// value, and st.hidden hides a blinking caret. A file input never paints a
// caret, even when focused.
func fieldSpanState(raw string, ctrl Control, st caretState, focused bool) string {
	kind := strings.ToLower(strings.TrimSpace(ctrl.Type))
	tag := strings.ToLower(strings.TrimSpace(ctrl.Tag))
	if kind == "file" {
		st.hidden = true
	}
	text := shownText(kind, ctrl.Value)
	ph := ""
	if text == "" && kind != "file" {
		ph = attrValue(raw, "placeholder")
	}
	if st.preedit != "" {
		ph = ""
	}

	n := runeLen(ctrl.Value)
	caret := st.caret
	if caret < 0 {
		caret = n
	}

	caret = clampPos(caret, n)
	start, end := rangeOf(st.start, st.end, n)
	whole := focused && (st.all || (start != end && start == 0 && end == n))

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
		if focused && !whole && !st.hidden {
			inner = caretSpan() + inner
		}
	case st.preedit != "":
		inner = markedPreedit(text, shownText(kind, st.preedit), st.start,
			st.end, st.preeditStart, st.preeditEnd)
	case whole:
		inner = selectionSpan(text)
	default:
		inner = markedText(text, focused, start, end, caret, st.hidden)
	}

	return openTagDrop("span", raw, extras, spanDrop) + inner + "</span>"
}

func caretSpan() string {
	return `<span data-ownframe-caret="1"></span>`
}
