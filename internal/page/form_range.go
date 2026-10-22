package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// bounds normalizes the form range to [start, end) clamped to n runes.
func (s *formState) bounds(n int) (int, int) {
	return form.RangeOf(s.anchor, s.caret, n)
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
	fs := p.formState()
	if fs == nil || fs.focusID == "" {
		p.clearRange()

		return
	}

	c, ok := fs.byID[fs.focusID]
	caret, anchor, action := form.ClampRange(form.FocusView{
		FocusID: fs.focusID,
		Caret:   fs.caret,
		Anchor:  fs.anchor,
		Control: c,
		Present: ok,
	})
	if action == form.RangeClear {
		p.clearRange()

		return
	}

	fs.caret, fs.anchor = caret, anchor
}
