package page

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"image"
	"sync"
	"sync/atomic"
	"time"
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

// runTick runs fn then the caret blink, then records in-place op patches
// as dirty rects. A no-op leaves FrameDirty unset so idle park can sleep.
func (p *Page) runTick(ctx context.Context, fn func(context.Context) error) error {
	gen := p.Generation()
	if fn == nil {
		err := p.caretBlink(ctx)
		p.flagTickDirty(gen)

		return err
	}

	snap := p.captureTickOps()
	if err := fn(ctx); err != nil {
		p.freeOps(snap)

		return err
	}

	err := p.caretBlink(ctx)
	p.noteTickDamage(gen, snap)
	p.freeOps(snap)

	return err
}

func (p *Page) captureTickOps() []DisplayOp {
	if p.draw.Display == nil {
		return nil
	}

	return append(p.rentOps(len(p.draw.Display.Ops)), p.draw.Display.Ops...)
}

func (p *Page) noteTickDamage(gen uint64, snap []DisplayOp) {
	if p.Generation() == gen && p.draw.Display != nil && snap != nil {
		before := *p.draw.Display
		before.Ops = snap
		d, full, _ := diffDisplay(&before, p.draw.Display)
		if full {
			p.markFull()
		} else {
			for i := 0; i < d.n; i++ {
				p.markRect(d.rects[i])
			}
		}
	}

	p.flagTickDirty(gen)
}

func (p *Page) flagTickDirty(gen uint64) {
	if p.Generation() != gen || p.dirty.Full || len(p.dirty.Rects) > 0 {
		p.frameDirty.Store(true)
	}
}

// Ticking reports whether a frame callback is registered. A Page stays on
// the buffer path: Tick marks mutated ops, and a no-op tick can idle.
func (p *Page) Ticking() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.fn != nil
}

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

var _ host.FrameWanter = (*Page)(nil)

// WantFrameAt declares the next time the page needs the window to draw. The
// desktop window reads it once per frame, before it draws, and schedules its
// idle wake for that time; a time at or before now asks for the next frame at
// once and keeps a per-frame animation on the vsync loop. The window clears the
// request when it reads it, like the FrameDirty opt-in, so a tick that wants
// another frame must ask again on every frame. The zero time means the page
// does not ask: the window falls back to its keep-alive pacing. Serve and the
// mobile views do not read it. WantFrameAt may be called from the tick or from
// an event handler.
func (p *Page) WantFrameAt(at time.Time) {
	if at.IsZero() {
		atomic.StoreInt64(&p.frameAt, 0)
		return
	}
	atomic.StoreInt64(&p.frameAt, at.UnixNano())
}

// WantFrameIn declares the next frame in d. d <= 0 asks for the next frame at
// once. It is WantFrameAt(time.Now().Add(d)) against the window's clock.
func (p *Page) WantFrameIn(d time.Duration) { p.WantFrameAt(time.Now().Add(d)) }

// NextFrameAt returns the deadline set by WantFrameAt or WantFrameIn and
// clears it. It reports the zero time when the page did not ask. The window
// calls it once per frame; the page does not.
func (p *Page) NextFrameAt() time.Time {
	n := atomic.SwapInt64(&p.frameAt, 0)
	if n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n)
}
