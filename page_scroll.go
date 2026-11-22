package ownframe

import (
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Scroll reads and sets offset, windowing, and the pinned layer.
type Scroll struct{ page *Page }

// Scroll returns offset, windowing, and pin controls for this page.
func (p *Page) Scroll() Scroll { return Scroll{page: p} }

// ScrollOffset returns the last offset the window reported.
func (s Scroll) ScrollOffset() (int, int) {
	p := innerPage(s.page)
	if p == nil {
		return 0, 0
	}

	return p.ScrollOffset()
}

// SetScrollOffset records the window offset. Tests drive the row window with it.
func (s Scroll) SetScrollOffset(x, y int) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetScrollOffset(x, y)
}

// SetWindowing opts into redraws on scroll changes.
func (s Scroll) SetWindowing(on bool) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetWindowing(on)
}

// Windowing reports whether scroll changes redraw this page.
func (s Scroll) Windowing() bool {
	p := innerPage(s.page)
	if p == nil {
		return false
	}

	return p.Windowing()
}

// SetAllowScroll turns window panning on or off.
func (s Scroll) SetAllowScroll(on bool) {
	p := innerPage(s.page)
	if p == nil {
		return
	}
	p.SetAllowScroll(on)
}

// AllowScroll reports that the window may pan this page.
func (s Scroll) AllowScroll() bool {
	p := innerPage(s.page)
	if p == nil {
		return false
	}

	return p.AllowScroll()
}

// SetViewportPinZ pins operations at or above z to the viewport.
func (s Scroll) SetViewportPinZ(z int) {
	inner := innerPage(s.page)
	if inner == nil {
		return
	}
	inner.SetViewportPinZ(z)
}

// SetScrollWindow installs a row-window callback.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetScrollWindow(fn)
}

// ScrollTo queues an absolute scroll move for the window.
func (p *Page) ScrollTo(x, y int) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.ScrollTo(x, y)
}

// ScrollBy queues a scroll delta for the window.
func (p *Page) ScrollBy(dx, dy int) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.ScrollBy(dx, dy)
}

// ScrollRequest is one queued window scroll request.
type ScrollRequest struct {
	X, Y     int
	Absolute bool
}

// ScrollbarMode controls whether a page shows scrollbar thumbs.
type ScrollbarMode = host.ScrollbarMode

const (
	// ScrollbarAuto follows the platform default: visible on desktop and web,
	// hidden on Android and iOS.
	ScrollbarAuto = host.ScrollbarAuto
	// ScrollbarVisible shows scrollbar thumbs on every platform.
	ScrollbarVisible = host.ScrollbarVisible
	// ScrollbarHidden hides scrollbar thumbs on every platform.
	ScrollbarHidden = host.ScrollbarHidden
)

// TouchScrollOptions tunes finger sensitivity and inertial scroll decay.
type TouchScrollOptions = page.TouchScrollOptions
