package ownframe

// Scroll is one queued window scroll request.
type Scroll struct {
	X, Y     int
	Absolute bool
}

// TakeScroll returns the pending scroll request and clears it.
// Telegram and spreadsheet tests drain host scroll this way.
func (p *Page) TakeScroll() (Scroll, bool) {
	req, ok := innerPage(p).TakeScroll()
	return Scroll{X: req.X, Y: req.Y, Absolute: req.Absolute}, ok
}

// StepScrollWindow prepares visible rows and reports whether they changed.
// css-preview drives the row window with it.
func (p *Page) StepScrollWindow() bool {
	return innerPage(p).StepScrollWindow()
}
