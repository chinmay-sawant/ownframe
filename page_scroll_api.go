package ownframe

// SetScrollWindow installs a row-window callback.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) {
	innerPage(p).SetScrollWindow(fn)
}
