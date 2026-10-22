package form

import "strings"

// Caret is the focus, caret, and range used to rewrite controls.
// A negative Caret means the end of the value.
type Caret struct {
	Focus        string
	Caret        int
	Start        int
	End          int
	All          bool
	Preedit      string
	PreeditStart int
	PreeditEnd   int
}

// FocusView is the focused control's caret, range, and preedit.
// Present is false when FocusID is not in the control map.
type FocusView struct {
	FocusID      string
	Caret        int
	Anchor       int
	All          bool
	Control      Control
	Present      bool
	PreeditID    string
	Preedit      string
	PreeditStart int
	PreeditEnd   int
}

// CaretOf builds the rewrite caret from the focused control.
// A missing focus or a non-editable control leaves Caret at -1.
func CaretOf(v FocusView) Caret {
	st := Caret{Caret: -1}
	if v.FocusID == "" || !v.Present {
		return st
	}

	st.Focus = v.Control.ID
	if !CanEdit(v.Control) {
		return st
	}

	n := RuneLen(v.Control.Value)
	st.Caret = ClampPos(v.Caret, n)
	st.Start, st.End = RangeOf(v.Anchor, v.Caret, n)
	st.All = v.All
	if v.PreeditID == v.Control.ID {
		st.Preedit = v.Preedit
		st.PreeditStart = v.PreeditStart
		st.PreeditEnd = v.PreeditEnd
	}

	return st
}

// CaretText is the painted field text and the caret's rune offset.
// ok is false when a preedit range is still open, so the caret stays hidden.
func CaretText(v FocusView) (string, int, bool) {
	c := v.Control
	kind := strings.ToLower(strings.TrimSpace(c.Type))
	if v.PreeditID != c.ID {
		return ShownText(kind, c.Value), ClampPos(v.Caret, RuneLen(c.Value)), true
	}

	if v.PreeditStart != v.PreeditEnd {
		return "", 0, false
	}

	base := []rune(c.Value)
	start, end := RangeOf(v.Anchor, v.Caret, len(base))
	value := string(base[:start]) + v.Preedit + string(base[end:])

	return ShownText(kind, value), start + RuneAtByte(v.Preedit, v.PreeditStart), true
}
