package ownframe

// SetScrollWindow installs a row-window callback.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) {
	innerPage(p).SetScrollWindow(fn)
}

// SetViewportPinZ pins operations at or above z to the viewport.
func (p *Page) SetViewportPinZ(z int) { innerPage(p).SetViewportPinZ(z) }

// ViewportPinZ is the pinned layer threshold.
func (p *Page) ViewportPinZ() int { return innerPage(p).ViewportPinZ() }

// ScrollOffset returns the last offset the window reported.
func (p *Page) ScrollOffset() (int, int) { return innerPage(p).ScrollOffset() }

// SetScrollOffset records the window offset. Tests drive the row window with it.
func (p *Page) SetScrollOffset(x, y int) { innerPage(p).SetScrollOffset(x, y) }

// SetWindowing opts into redraws on scroll changes.
func (p *Page) SetWindowing(on bool) { innerPage(p).SetWindowing(on) }

// Windowing reports whether scroll changes redraw this page.
func (p *Page) Windowing() bool { return innerPage(p).Windowing() }

// ScrollTo queues an absolute scroll move for the window.
func (p *Page) ScrollTo(x, y int) { innerPage(p).ScrollTo(x, y) }

// ScrollBy queues a scroll delta for the window.
func (p *Page) ScrollBy(dx, dy int) { innerPage(p).ScrollBy(dx, dy) }

// SetAllowScroll turns window panning on or off.
func (p *Page) SetAllowScroll(on bool) { innerPage(p).SetAllowScroll(on) }

// AllowScroll reports that the window may pan this page.
func (p *Page) AllowScroll() bool { return innerPage(p).AllowScroll() }
