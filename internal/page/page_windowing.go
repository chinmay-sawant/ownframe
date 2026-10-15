package page

// windowingState is the scroll offset the window last reported and the
// row-window callback. The window owns scrolling; the callback slices app
// rows before each draw so only visible rows become boxes and operations.
// Writer: the window via SetWindowing, SetScrollOffset, SetScrollWindow.
// Lock: the page loop serializes all writes; readers run on the same loop.
type windowingState struct {
	x  int
	y  int
	on bool
	fn func(offsetY, viewH int) bool
}

// SetWindowing opts into redraws on scroll changes. Leave it off and the
// window scrolls by blitting, which is cheaper for fully laid-out pages.
func (p *Page) SetWindowing(on bool) { p.windowing.on = on }

// Windowing reports whether scroll changes redraw this page.
func (p *Page) Windowing() bool { return p.windowing.on }

// ScrollOffset returns the last offset the window reported, in CSS pixels.
func (p *Page) ScrollOffset() (int, int) { return p.windowing.x, p.windowing.y }

// SetScrollOffset records the window offset. The window calls it when the
// offset changes; tests call it to drive the row window headlessly.
func (p *Page) SetScrollOffset(x, y int) { p.windowing.x, p.windowing.y = x, y }

// SetScrollWindow installs a row-window callback. It returns true after
// changing data, false when the existing overscan covers the viewport.
// The host skips scroll relayout when false. Redraw also calls it first.
func (p *Page) SetScrollWindow(fn func(offsetY, viewH int) bool) { p.windowing.fn = fn }

// applyScrollWindow runs the row-window callback before the template
// executes. A nil callback, or a page that opted out, draws unchanged.
func (p *Page) applyScrollWindow() {
	if !p.windowing.on || p.windowing.fn == nil {
		return
	}

	p.windowing.fn(p.windowing.y, p.height)
}
