package ownframe

// SetScrollWindow installs a row-window callback.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) {
	innerPage(p).SetScrollWindow(fn)
}

// ScrollTo queues an absolute scroll move for the window.
func (page *Page) ScrollTo(x, y int) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.ScrollTo(x, y)
}

// ScrollBy queues a scroll delta for the window.
func (page *Page) ScrollBy(dx, dy int) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.ScrollBy(dx, dy)
}
