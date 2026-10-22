package ownframe

// Size returns the frame size in CSS pixels.
func (p *Page) Size() (int, int) { return innerPage(p).Size() }

// MinSize returns the configured desktop minimum.
func (p *Page) MinSize() (int, int) { return innerPage(p).MinSize() }

// MaxSize returns the largest window the page asks its host for.
func (p *Page) MaxSize() (int, int) { return innerPage(p).MaxSize() }

// Clamp applies the desktop minimum or the native mobile viewport.
func (p *Page) Clamp(width, height int) (int, int) {
	return innerPage(p).Clamp(width, height)
}

// SetSize stores the frame size used by the next Redraw.
func (p *Page) SetSize(width, height int) {
	innerPage(p).SetSize(width, height)
}
