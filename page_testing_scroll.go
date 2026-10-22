package ownframe

// TakeScroll returns the pending scroll request and clears it.
func (t Testing) TakeScroll() (Scroll, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Scroll{}, false
	}
	req, ok := p.TakeScroll()
	return Scroll{X: req.X, Y: req.Y, Absolute: req.Absolute}, ok
}

// StepScrollWindow prepares visible rows and reports whether they changed.
func (t Testing) StepScrollWindow() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.StepScrollWindow()
}
