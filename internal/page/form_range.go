package page

import "unicode/utf8"

// bounds normalizes the form range to [start, end) clamped to n runes.
func (s *formState) bounds(n int) (int, int) {
	return rangeOf(s.anchor, s.caret, n)
}

// rangeOf normalizes two offsets into [start, end) clamped to n runes.
func rangeOf(a, b, n int) (int, int) {
	start, end := a, b
	if start > end {
		start, end = end, start
	}

	start = clampPos(start, n)
	end = clampPos(end, n)
	if end < start {
		end = start
	}

	return start, end
}

func clampPos(pos, n int) int {
	switch {
	case pos < 0:
		return 0
	case pos > n:
		return n
	default:
		return pos
	}
}

func runeLen(s string) int {
	return utf8.RuneCountInString(s)
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
