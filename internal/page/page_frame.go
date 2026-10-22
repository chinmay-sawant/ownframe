package page

import "image"

// MarkRect adds a CSS-pixel region to the dirty rect. A tick uses it for a
// line that has no element id, so the id is not what a click hits.
func (p *Page) MarkRect(r image.Rectangle) { p.markRect(r) }

// UseFrameDirty is a kept alias. Tick sets the flag when ops change or dirty
// rects remain. An explicit call still sets it, including a no-op tick that
// wants the buffer path without damage.
func (p *Page) UseFrameDirty() { p.frameDirty.Store(true) }

// FrameDirty reports the opt-in from UseFrameDirty or from a mutating tick
// and clears it.
func (p *Page) FrameDirty() bool { return p.frameDirty.Swap(false) }
