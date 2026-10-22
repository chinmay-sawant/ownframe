package page

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// SetAllowScroll turns window panning on or off.
// The pad screen turns it off so a drag moves the TV pointer.
func (p *Page) SetAllowScroll(on bool) {
	if p == nil {
		return
	}

	p.allowScroll = on
}

// AllowScroll reports that the window may pan this page.
func (p *Page) AllowScroll() bool {
	if p == nil {
		return true
	}

	return p.allowScroll
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
func (p *Page) SetScrollOffset(x, y int) { p.scroll.SetOffset(x, y) }

// SetScrollWindow installs a row-window callback. It returns true after
// changing data, false when the existing overscan covers the viewport.
// The host skips scroll relayout when false. Redraw also calls it first.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) { p.scroll.SetFn(fn) }

// applyScrollWindow runs the row-window callback before the template
// executes. A nil callback, or a page that opted out, draws unchanged.
func (p *Page) applyScrollWindow() { p.scroll.Apply(p.height) }

// SetViewportPinZ pins the page's top layers. Operations whose z-index is at
// least z are drawn at the offset the display was built with, so they stay at
// the viewport while the page scrolls and need no redraw. Zero disables the
// pinning and draws every operation with the scroll, as before.
func (p *Page) SetViewportPinZ(z int) { p.pinZ = z }

// ViewportPinZ is the pinned layer threshold.
func (p *Page) ViewportPinZ() int { return p.pinZ }
