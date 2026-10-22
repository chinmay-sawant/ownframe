package form

// DeleteWordValue removes the selection, or the word before the caret.
func DeleteWordValue(c Control, caret, anchor int) (Control, int, int, bool) {
	runes := []rune(c.Value)
	start, end := RangeOf(anchor, caret, len(runes))
	if start != end {
		return SetRunes(c, append(runes[:start], runes[end:]...)), start, start, true
	}

	from := PrevWordStart(c.Value, start)
	if from == start {
		return c, caret, anchor, false
	}

	return SetRunes(c, append(runes[:from], runes[start:]...)), from, from, true
}
