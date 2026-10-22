package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

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
func (p *Page) caretOf() form.Caret {
	fs := p.formState()
	if fs == nil || fs.focusID == "" {
		return form.CaretOf(form.FocusView{})
	}

	c, ok := fs.byID[fs.focusID]

	return form.CaretOf(form.FocusView{
		FocusID:      fs.focusID,
		Caret:        fs.caret,
		Anchor:       fs.anchor,
		All:          fs.all,
		Control:      c,
		Present:      ok,
		PreeditID:    fs.preeditID,
		Preedit:      fs.preedit,
		PreeditStart: fs.preeditStart,
		PreeditEnd:   fs.preeditEnd,
	})
}
