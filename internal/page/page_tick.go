package page

import (
	"context"
	"sync"
)

// tickSlot holds the frame callback. SetTick can run on any goroutine while
// Tick runs on the page loop, so the pointer is guarded. The callback itself
// runs without the lock, so it may call SetTick.
type tickSlot struct {
	mu sync.Mutex
	fn func(ctx context.Context) error
}

// SetTick registers fn to run once per frame before the window draws. fn can
// read the display list and change an operation's geometry or text, or call
// Redraw; either changes the next frame. nil removes the callback.
// Serve does not tick. Tick marks patched ops dirty; callers need not call
// UseFrameDirty.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	p.mu.Lock()
	p.fn = fn
	p.mu.Unlock()
}

// Tick runs the frame callback registered with SetTick and the caret blink.
// With no callback and no focused field it does nothing. A callback that
// mutated ops sets FrameDirty and TakeDirty; a no-op tick stays clean.
func (p *Page) Tick(ctx context.Context) error {
	p.mu.Lock()
	fn := p.fn
	p.mu.Unlock()

	return p.runTick(ctx, fn)
}
