package page

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
	"image"
	"sync"
	"sync/atomic"
	"time"
)

// tickSlot holds the frame callback. SetTick can run on any goroutine while
// Tick runs on the page loop, so the pointer is guarded. The callback itself
// runs without the lock, so it may call SetTick.
type tickSlot struct {
	mu    sync.Mutex
	fn    func(ctx context.Context) error
	hasFn atomic.Bool
}

// SetTick registers fn to run once per frame before the window draws. fn can
// read the display list and change an operation's geometry or text, or call
// Redraw; either changes the next frame. nil removes the callback.
// Serve does not tick. Tick marks patched ops dirty; callers need not call
// UseFrameDirty.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	p.mu.Lock()
	p.fn = fn
	p.hasFn.Store(fn != nil)
	p.mu.Unlock()
	notifyFrame(p)
}

// Tick steps the automatic CSS motion driver, runs the frame callback
// registered with SetTick, and blinks the caret. With no callback, no focused
// field, and no animation it does nothing. A mutating tick sets FrameDirty and
// TakeDirty; a no-op tick stays clean.
func (p *Page) Tick(ctx context.Context) error {
	p.tickRunning.Store(true)
	defer p.tickRunning.Store(false)
	var fn func(context.Context) error
	if p.hasFn.Load() {
		p.mu.Lock()
		fn = p.fn
		p.mu.Unlock()
	}

	return runTick(p, ctx, fn)
}

// runTick runs the CSS layout pass, steps the paint driver, runs fn, then the
// caret blink, then records in-place op patches as dirty rects. The layout pass
// runs first so every later sample composes against the display it produced.
// The driver damage is marked before the callback snapshot, so the callback
// diff never covers an animation change. A no-op leaves FrameDirty unset so
// idle park can sleep. The 3D re-projection runs last, so it cannot invalidate
// the callback's snapshot.
func runTick(p *Page, ctx context.Context, fn func(context.Context) error) error {
	gen := p.Generation()
	stepMotion(p, ctx)
	if fn == nil {
		err := caretBlink(p, ctx)
		flagTickDirty(p, gen)
		if err != nil {
			return err
		}

		return restampTick(p, ctx, gen)
	}

	p.tickMutated.Store(false)
	p.tickActive = true
	p.tickSnap = nil

	err := fn(ctx)
	p.tickActive = false
	if err != nil {
		if p.tickSnap != nil {
			freeOps(p, p.tickSnap)
			p.tickSnap = nil
		}

		return err
	}

	blinkErr := caretBlink(p, ctx)

	mutated := p.tickMutated.Load() || p.Generation() != gen || p.dirty.Full || len(p.dirty.Rects) > 0 || p.tickSnap != nil
	if !mutated {
		if p.restampRedraw {
			return restampTick(p, ctx, gen)
		}

		return blinkErr
	}

	snap := p.tickSnap
	p.tickSnap = nil
	if snap != nil {
		noteTickDamage(p, gen, snap)
		freeOps(p, snap)
	} else {
		flagTickDirty(p, gen)
	}

	if blinkErr != nil {
		return blinkErr
	}

	return restampTick(p, ctx, gen)
}

// restampTick applies the 3D re-projection requests of the sample stepMotion
// just took, then performs the one redraw that turns retained layout state on
// when the pass needed it. Both replace the display the tick started with, so
// the tick is marked dirty even when the callback changed nothing.
func restampTick(p *Page, ctx context.Context, gen uint64) error {
	restampMotion(p, ctx)
	if p.restampRedraw {
		// The debt stays set through the redraw: the bind that queued it
		// already turned retainState on, so the frame key alone would call
		// this redraw unchanged and skip the retained layout it owes.
		if err := p.Redraw(ctx); err != nil {
			return err
		}
		p.restampRedraw = false
	}
	flagTickDirty(p, gen)

	return nil
}

func captureTickOps(p *Page) *opBuffer {
	if p.draw.Display == nil {
		return nil
	}

	buf := rentOps(p, len(p.draw.Display.Ops))
	buf.Ops = append(buf.Ops, p.draw.Display.Ops...)
	return buf
}

func noteTickDamage(p *Page, gen uint64, snap *opBuffer) {
	if p.Generation() == gen && p.draw.Display != nil && snap != nil {
		if p.motion != nil && p.motion.Active() {
			p.motion.CheckManualOverlap(snap.Ops, p.draw.Display.Ops)
		}
		before := *p.draw.Display
		before.Ops = snap.Ops
		d, full, _ := diffDisplay(&before, p.draw.Display)
		if full {
			markFull(p)
		} else {
			for i := 0; i < d.n; i++ {
				markRect(p, d.rects[i])
			}
		}
	}

	flagTickDirty(p, gen)
}

func flagTickDirty(p *Page, gen uint64) {
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
func (p *Page) MarkRect(r image.Rectangle) { markRect(p, r) }

// UseFrameDirty is a kept alias. Tick sets the flag when ops change or dirty
// rects remain. An explicit call still sets it, including a no-op tick that
// wants the buffer path without damage.
func (p *Page) UseFrameDirty() {
	defer notifyFrame(p)
	p.tickMutated.Store(true)
	p.frameDirty.Store(true)
}

// FrameDirty reports the opt-in from UseFrameDirty or from a mutating tick
// and clears it.
func (p *Page) FrameDirty() bool { return p.frameDirty.Swap(false) }

var _ host.FrameWanter = (*Page)(nil)

// WantFrameAt declares the next time the page needs the window to draw. The
// desktop window reads it once per frame, before it draws, and schedules its
// idle wake for that time; a time at or before now asks for the next frame at
// once and keeps a per-frame animation on the active drawing loop. The window clears the
// request when it reads it, like the FrameDirty opt-in, so a tick that wants
// another frame must ask again on every frame. The zero time means the page
// does not ask: a static page waits for events, while a callback or file watch
// retains compatibility polling. Serve and the
// mobile views do not read it. WantFrameAt may be called from the tick or from
// an event handler.
func (p *Page) WantFrameAt(at time.Time) {
	defer notifyFrame(p)
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
	var at time.Time
	if n != 0 {
		at = time.Unix(0, n)
	}
	if blink := caretDeadline(p); !blink.IsZero() && (at.IsZero() || blink.Before(at)) {
		at = blink
	}
	return at
}

// frameState is the page's per-frame opt-ins: the dirty-rect opt-in the
// window reads, and the next-frame deadline a page sets to pace an idle
// window. Both clear when the window reads them. Page embeds this state to
// group the atomic handoff fields.
// Writer: the page loop sets, the window take-clears (page_tick.go).
// Lock: none, sync/atomic carries the handoff; no mutex needed.
//
// frameAt is stored as Unix nanoseconds and accessed with sync/atomic so a
// frame deadline can be handed off between the tick, handlers, and the game
// loop without a lock.
type frameState struct {
	frameDirty  atomic.Bool
	frameAt     int64 // atomic: Unix nanoseconds
	frameOnce   sync.Once
	frameEvents chan struct{}
	tickRunning atomic.Bool
}

var _ host.FrameScheduler = (*Page)(nil)

// FrameEvents coalesces notifications without blocking a producer. The host
// consumes these events; no page data is read on its timer goroutine.
func (p *Page) FrameEvents() <-chan struct{} {
	p.frameOnce.Do(func() { p.frameEvents = make(chan struct{}, 1) })
	return p.frameEvents
}

func notifyFrame(p *Page) {
	if p.tickRunning.Load() {
		return
	}
	p.FrameEvents()
	select {
	case p.frameEvents <- struct{}{}:
	default:
	}
}

// FramePolling preserves callbacks that discover changes by polling and the
// existing file reload contract. Static pages need neither keep-alive.
func (p *Page) FramePolling() bool { return p.hasFn.Load() || p.Watching() }

func caretDeadline(p *Page) time.Time {
	c, ok := focusedEditable(p)
	fs := formState(p)
	if !ok || fs == nil || form.IsFileInput(c) {
		return time.Time{}
	}
	start, end := form.RangeOf(fs.Anchor, fs.Caret, runeLen(c.Value))
	if fs.All || start != end {
		return time.Time{}
	}
	if fs.BlinkAt.IsZero() {
		return time.Now().Add(form.BlinkInterval)
	}
	return fs.BlinkAt.Add(form.BlinkInterval)
}
