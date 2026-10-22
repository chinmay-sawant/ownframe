package page

// bounds normalizes the form range to [start, end) clamped to n runes.
func (s *formState) bounds(n int) (int, int) {
	return rangeOf(s.anchor, s.caret, n)
}

// clearRange drops the caret and the anchor.
func (p *Page) clearRange() {
	if p.formState() == nil {
		return
	}

	p.formState().caret, p.formState().anchor, p.formState().all = 0, 0, false
}

// clampRange clamps the caret and the anchor to the focused value after a
// merge. A vanished focus clears them.
func (p *Page) clampRange() {
	if p.formState() == nil || p.formState().focusID == "" {
		p.clearRange()

		return
	}

	c, ok := p.formState().byID[p.formState().focusID]
	if !ok || !canEdit(c) {
		p.formState().caret, p.formState().anchor = 0, 0

		return
	}

	n := runeLen(c.Value)
	p.formState().caret = clampPos(p.formState().caret, n)
	p.formState().anchor = clampPos(p.formState().anchor, n)
}
