package ownframe

import "context"

// Tick runs the frame callback registered with SetTick and the caret blink.
func (p *Page) Tick(ctx context.Context) error {
	return innerPage(p).Tick(ctx)
}

// Ticking reports whether a frame callback is registered.
func (p *Page) Ticking() bool { return innerPage(p).Ticking() }
