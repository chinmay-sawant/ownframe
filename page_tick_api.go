package ownframe

// Ticking reports whether a frame callback is registered.
func (p *Page) Ticking() bool { return innerPage(p).Ticking() }
