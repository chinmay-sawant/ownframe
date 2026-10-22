package page

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
