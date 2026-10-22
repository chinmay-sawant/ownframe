package ownframe

import "image"

// MarkRect adds a CSS-pixel region to the dirty rect from a tick.
func (t Testing) MarkRect(r image.Rectangle) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.MarkRect(r)
}

// UseFrameDirty opts the next window draw into the buffer path.
func (t Testing) UseFrameDirty() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.UseFrameDirty()
}

// FrameDirty reports and clears the frame-dirty opt-in.
func (t Testing) FrameDirty() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.FrameDirty()
}
