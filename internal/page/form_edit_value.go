package page

// insertValue inserts text at the caret and replaces the selection.
func (p *Page) insertValue(c Control, caret, anchor int, text string) (Control, int, int, bool) {
	if text == "" {
		return c, caret, anchor, false
	}

	runes := []rune(c.Value)
	start, end := rangeOf(anchor, caret, len(runes))
	ins := []rune(text)
	// GEN-28: keep the merged value inside the field cap.
	if limit := fieldCap(c) - (len(runes) - (end - start)); limit < len(ins) {
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

	return setRunes(c, out), pos, pos, true
}

// backspaceValue removes the selection, or the rune before the caret.
func (p *Page) backspaceValue(c Control, caret, anchor int) (Control, int, int, bool) {
	runes := []rune(c.Value)
	start, end := rangeOf(anchor, caret, len(runes))
	if start != end {
		return setRunes(c, append(runes[:start], runes[end:]...)), start, start, true
	}

	if start == 0 {
		return c, caret, anchor, false
	}

	return setRunes(c, append(runes[:start-1], runes[start:]...)), start - 1, start - 1, true
}

// deleteWordValue removes the selection, or the word before the caret.
func (p *Page) deleteWordValue(c Control, caret, anchor int) (Control, int, int, bool) {
	runes := []rune(c.Value)
	start, end := rangeOf(anchor, caret, len(runes))
	if start != end {
		return setRunes(c, append(runes[:start], runes[end:]...)), start, start, true
	}

	from := prevWordStart(c.Value, start)
	if from == start {
		return c, caret, anchor, false
	}

	return setRunes(c, append(runes[:from], runes[start:]...)), from, from, true
}

func setRunes(c Control, runes []rune) Control {
	c.Value = string(runes)

	return c
}
