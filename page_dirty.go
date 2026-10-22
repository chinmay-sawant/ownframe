package ownframe

import "image"

// MarkRect adds a CSS-pixel region to the dirty rect from a tick.
func (p *Page) MarkRect(r image.Rectangle) { innerPage(p).MarkRect(r) }

// UseFrameDirty opts the next window draw into the buffer path.
func (p *Page) UseFrameDirty() { innerPage(p).UseFrameDirty() }

// FrameDirty reports and clears the frame-dirty opt-in.
func (p *Page) FrameDirty() bool { return innerPage(p).FrameDirty() }
