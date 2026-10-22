package page

// Ticking reports whether a frame callback is registered. A Page stays on
// the buffer path: Tick marks mutated ops, and a no-op tick can idle.
func (p *Page) Ticking() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.fn != nil
}
