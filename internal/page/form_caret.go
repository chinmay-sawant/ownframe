package page

// caretState is the focus, caret, and range the control rewrite paints.
// A negative caret means the end of the value.
type caretState struct {
	focus        string
	caret        int
	start        int
	end          int
	all          bool
	hidden       bool
	preedit      string
	preeditStart int
	preeditEnd   int
}

// setCaret moves the caret to pos, keeping the anchor when extend is true.
func (p *Page) setCaret(pos int, extend bool) {
	if p.form == nil {
		return
	}

	pos = clampPos(pos, p.focusedLen())
	if !extend {
		p.form.anchor = pos
	}

	p.form.caret = pos
	p.form.all = false
	p.resetCaretBlink()
}

func (p *Page) focusedLen() int {
	c, ok := p.focusedEditable()
	if !ok {
		return 0
	}

	return runeLen(c.Value)
}

// caretEnd puts the caret and the anchor at the end of c.
func (p *Page) caretEnd(c Control) {
	n := runeLen(c.Value)
	p.form.caret, p.form.anchor = n, n
	p.form.all = false
	p.resetCaretBlink()
}

// hasSelection reports whether the focused field has a non-empty range.
func (p *Page) hasSelection() bool {
	c, ok := p.typingTarget()
	if !ok {
		return false
	}

	start, end := p.form.bounds(runeLen(c.Value))

	return start != end
}

// caretOf returns the state the control rewrite paints.
func (p *Page) caretOf() caretState {
	st := caretState{caret: -1}
	if p.form == nil || p.form.focusID == "" {
		return st
	}

	c, ok := p.form.byID[p.form.focusID]
	if !ok {
		return st
	}

	st.focus = c.ID
	if !canEdit(c) {
		return st
	}

	n := runeLen(c.Value)
	st.caret = clampPos(p.form.caret, n)
	st.start, st.end = p.form.bounds(n)
	st.all = p.form.all
	st.hidden = !p.form.blinkOn || isFileInput(c)
	if p.form.preeditID == c.ID {
		st.preedit = p.form.preedit
		st.preeditStart = p.form.preeditStart
		st.preeditEnd = p.form.preeditEnd
	}

	return st
}
