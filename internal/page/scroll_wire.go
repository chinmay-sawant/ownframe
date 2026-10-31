package page

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page/scroll"
)

// TouchScrollOptions tunes touch scrolling. Sensitivity defaults to 1 and is
// limited to 0.25-3. Deceleration defaults to 5; larger values stop sooner.
type TouchScrollOptions = scroll.Touch

// SetAllowScroll turns window panning on or off.
// The pad screen turns it off so a drag moves the TV pointer.
func (p *Page) SetAllowScroll(on bool) {
	if p == nil {
		return
	}
	p.scroll.SetAllow(on)
}

// AllowScroll reports that the window may pan this page.
func (p *Page) AllowScroll() bool {
	if p == nil {
		return true
	}
	return p.scroll.Allow()
}

// ScrollTo queues an absolute scroll move for the window. The window clamps
// the target to the content size and owns the offset.
func (p *Page) ScrollTo(x, y int) { p.scroll.To(x, y) }

// ScrollBy queues a scroll delta for the window. A second delta before the
// window consumes it adds to the first, and a delta after ScrollTo moves the
// pending target.
func (p *Page) ScrollBy(dx, dy int) { p.scroll.By(dx, dy) }

// TakeScroll returns the queued request and clears it.
func (p *Page) TakeScroll() (host.Scroll, bool) { return p.scroll.Take() }

// StepScrollWindow prepares the visible rows and reports whether they changed.
// A windowing page without a callback keeps the ordinary scroll redraw path.
func (p *Page) StepScrollWindow() bool {
	return p.scroll.Step(p.height)
}

// SetWindowing opts into redraws on scroll changes. Leave it off and the
// window scrolls by blitting, which is cheaper for fully laid-out pages.
func (p *Page) SetWindowing(on bool) { p.scroll.SetOn(on) }

// Windowing reports whether scroll changes redraw this page.
func (p *Page) Windowing() bool { return p.scroll.On() }

// ScrollOffset returns the last offset the window reported, in CSS pixels.
func (p *Page) ScrollOffset() (int, int) { return p.scroll.Offset() }

// SetScrollOffset records the window offset. The window calls it when the
// offset changes; tests call it to drive the row window headlessly.
func (p *Page) SetScrollOffset(x, y int) {
	p.scroll.SetOffset(x, y)
	p.formDirty = true
}

// SetScrollWindow installs a row-window callback. It returns true after
// changing data, false when the existing overscan covers the viewport.
// The host skips scroll relayout when false. Redraw also calls it first.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) { p.scroll.SetFn(fn) }

// applyScrollWindow runs the row-window callback before the template
// executes. A nil callback, or a page that opted out, draws unchanged.
func (p *Page) applyScrollWindow() {
	if p.scroll.Apply(p.height) {
		p.formDirty = true
	}
}

// SetViewportPinZ pins the page's top layers. Operations whose z-index is at
// least z are drawn at the offset the display was built with, so they stay at
// the viewport while the page scrolls and need no redraw. Zero disables the
// pinning and draws every operation with the scroll, as before.
func (p *Page) SetViewportPinZ(z int) { p.scroll.SetPinZ(z) }

// ViewportPinZ is the pinned layer threshold.
func (p *Page) ViewportPinZ() int { return p.scroll.PinZ() }

// ScrollbarMode reports the configured page thumb visibility mode.
func (p *Page) ScrollbarMode() host.ScrollbarMode {
	if p == nil {
		return host.ScrollbarAuto
	}
	return p.scroll.Bar()
}

// SetTouchScrollOptions changes touch scrolling for this page.
func (p *Page) SetTouchScrollOptions(options TouchScrollOptions) {
	if p == nil {
		return
	}
	p.scroll.SetTouch(options)
}

// TouchScrollSensitivity is the configured finger movement multiplier.
func (p *Page) TouchScrollSensitivity() float64 {
	if p == nil {
		return (*scroll.State)(nil).Sensitivity()
	}
	return p.scroll.Sensitivity()
}

// TouchScrollDeceleration is the configured fling decay per second.
func (p *Page) TouchScrollDeceleration() float64 {
	if p == nil {
		return (*scroll.State)(nil).Deceleration()
	}
	return p.scroll.Deceleration()
}

// TouchScrollOptions returns the active touch scroll settings.
func (p *Page) TouchScrollOptions() TouchScrollOptions {
	return TouchScrollOptions{
		Sensitivity:  p.TouchScrollSensitivity(),
		Deceleration: p.TouchScrollDeceleration(),
	}
}
