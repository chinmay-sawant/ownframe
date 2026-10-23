package ownframe

// Scroll reads and sets offset, windowing, and the pinned layer.
type Scroll struct{ page *Page }

// Scroll returns offset, windowing, and pin controls for this page.
func (p *Page) Scroll() Scroll { return Scroll{page: p} }

// ScrollOffset returns the last offset the window reported.
func (s Scroll) ScrollOffset() (int, int) {
	p := innerPage(s.page)
	if p == nil {
		return 0, 0
	}

	return p.ScrollOffset()
}

// SetScrollOffset records the window offset. Tests drive the row window with it.
func (s Scroll) SetScrollOffset(x, y int) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetScrollOffset(x, y)
}

// SetWindowing opts into redraws on scroll changes.
func (s Scroll) SetWindowing(on bool) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetWindowing(on)
}

// Windowing reports whether scroll changes redraw this page.
func (s Scroll) Windowing() bool {
	p := innerPage(s.page)
	if p == nil {
		return false
	}

	return p.Windowing()
}

// SetAllowScroll turns window panning on or off.
func (s Scroll) SetAllowScroll(on bool) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetAllowScroll(on)
}

// AllowScroll reports that the window may pan this page.
func (s Scroll) AllowScroll() bool {
	p := innerPage(s.page)
	if p == nil {
		return false
	}

	return p.AllowScroll()
}

// SetViewportPinZ pins operations at or above z to the viewport.
func (s Scroll) SetViewportPinZ(z int) { innerPage(s.page).SetViewportPinZ(z) }

// ViewportPinZ is the pinned layer threshold.
func (s Scroll) ViewportPinZ() int { return innerPage(s.page).ViewportPinZ() }
