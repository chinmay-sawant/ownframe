package form

// InsertValue inserts text at the caret and replaces the selection.
func InsertValue(c Control, caret, anchor int, text string) (Control, int, int, bool) {
	if text == "" {
		return c, caret, anchor, false
	}

	runes := []rune(c.Value)
	start, end := RangeOf(anchor, caret, len(runes))
	ins := []rune(text)
	// GEN-28: keep the merged value inside the field cap.
	if limit := FieldCap(c) - (len(runes) - (end - start)); limit < len(ins) {
		if limit <= 0 {
			return c, caret, anchor, false
		}
		ins = ins[:limit]
	}
	out := make([]rune, 0, len(runes)-(end-start)+len(ins))
	out = append(out, runes[:start]...)
	out = append(out, ins...)
	out = append(out, runes[end:]...)
	pos := start + len(ins)

	return SetRunes(c, out), pos, pos, true
}

// BackspaceValue removes the selection, or the rune before the caret.
func BackspaceValue(c Control, caret, anchor int) (Control, int, int, bool) {
	runes := []rune(c.Value)
	start, end := RangeOf(anchor, caret, len(runes))
	if start != end {
		return SetRunes(c, append(runes[:start], runes[end:]...)), start, start, true
	}

	if start == 0 {
		return c, caret, anchor, false
	}

	return SetRunes(c, append(runes[:start-1], runes[start:]...)), start - 1, start - 1, true
}

// SetRunes stores runes as c.Value.
func SetRunes(c Control, runes []rune) Control {
	c.Value = string(runes)

	return c
}

// ClampEdited caps c.Value to the field cap and pulls the caret and the
// anchor back inside the capped value.
func ClampEdited(c Control, caret, anchor int) (Control, int, int) {
	c.Value = ClampField(c.Value, FieldCap(c))
	if n := RuneLen(c.Value); caret > n || anchor > n {
		caret, anchor = min(caret, n), min(anchor, n)
	}

	return c, caret, anchor
}
