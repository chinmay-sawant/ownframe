package page

// caretState carries the focus, caret, and range used to rewrite controls.
// A negative caret means the end of the value.
type caretState struct {
	focus        string
	caret        int
	start        int
	end          int
	all          bool
	preedit      string
	preeditStart int
	preeditEnd   int
}

// setCaret moves the caret to pos, keeping the anchor when extend is true.
func (p *Page) setCaret(pos int, extend bool) {
	if p.formState() == nil {
		return
	}

	pos = clampPos(pos, p.focusedLen())
	if !extend {
		p.formState().anchor = pos
	}

	p.formState().caret = pos
	p.formState().all = false
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
	p.formState().caret, p.formState().anchor = n, n
	p.formState().all = false
	p.resetCaretBlink()
}

// hasSelection reports whether the focused field has a non-empty range.
func (p *Page) hasSelection() bool {
	c, ok := p.typingTarget()
	if !ok {
		return false
	}

	start, end := p.formState().bounds(runeLen(c.Value))

	return start != end
}

// caretOf returns the state the control rewrite paints.
func (p *Page) caretOf() caretState {
	st := caretState{caret: -1}
	if p.formState() == nil || p.formState().focusID == "" {
		return st
	}

	c, ok := p.formState().byID[p.formState().focusID]
	if !ok {
		return st
	}

	st.focus = c.ID
	if !canEdit(c) {
		return st
	}

	n := runeLen(c.Value)
	st.caret = clampPos(p.formState().caret, n)
	st.start, st.end = p.formState().bounds(n)
	st.all = p.formState().all
	if p.formState().preeditID == c.ID {
		st.preedit = p.formState().preedit
		st.preeditStart = p.formState().preeditStart
		st.preeditEnd = p.formState().preeditEnd
	}

	return st
}
