package page

import "image"

// MarkRect adds a CSS-pixel region to the dirty rect. A tick uses it for a
// line that has no element id, so the id is not what a click hits.
func (p *Page) MarkRect(r image.Rectangle) { p.markRect(r) }

// UseFrameDirty opts this tick into the partial replay. The window then
// repaints the rect from TakeDirty, or blits its buffer when that rect is
// empty. A tick that never calls it keeps the full replay.
func (p *Page) UseFrameDirty() { p.frameDirty.Store(true) }

// FrameDirty reports the opt-in from UseFrameDirty and clears it.
func (p *Page) FrameDirty() bool { return p.frameDirty.Swap(false) }
