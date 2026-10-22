package ownframe

// SetScrollWindow installs a row-window callback.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) {
	innerPage(p).SetScrollWindow(fn)
}

// SetViewportPinZ pins operations at or above z to the viewport.
func (p *Page) SetViewportPinZ(z int) { innerPage(p).SetViewportPinZ(z) }

// ViewportPinZ is the pinned layer threshold.
func (p *Page) ViewportPinZ() int { return innerPage(p).ViewportPinZ() }
