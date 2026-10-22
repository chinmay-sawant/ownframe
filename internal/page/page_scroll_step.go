package page

// StepScrollWindow prepares the visible rows and reports whether they changed.
// A windowing page without a callback keeps the ordinary scroll redraw path.
func (p *Page) StepScrollWindow() bool {
	return p.scroll.Step(p.height)
}
