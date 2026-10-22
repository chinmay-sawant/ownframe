package page

// Ticking reports whether a frame callback is registered. The window keeps a
// page with one on the full replay path, because a callback changes the
// display list in place and leaves no dirty rect behind.
func (p *Page) Ticking() bool {
	p.tickSlot.mu.Lock()
	defer p.tickSlot.mu.Unlock()

	return p.tickSlot.fn != nil
}
