package page

// afterRedrawState holds the after-redraw hook and the guards around it.
type afterRedrawState struct {
	fn     func()
	busy   bool
	ask    bool
	served bool
}

// SetAfterRedraw registers fn to run after a redraw commits a new display.
//
// The callback runs on the page loop, at the end of the commit, so the new
// display and the new hit boxes are installed and the app can read them. An
// app that keeps its state in the parsed tree uses it to reapply that state
// after the template replaced the tree.
//
// The callback cannot deadlock: it runs after the render cache released its
// lock, so an edit of the tree does not wait on itself. It cannot re-enter a
// redraw either: a Redraw asked for inside the callback waits for the commit
// to finish, and the page draws again once the callback returned.
//
// nil removes the callback. Register it from the page loop, like SetTick.
func (p *Page) SetAfterRedraw(fn func()) {
	if p == nil {
		return
	}

	p.afterRedraw.fn = fn
}

// runAfterRedraw calls the registered callback after a commit installed the
// display and the boxes. A callback that asks for a redraw does not re-enter
// the commit: the request waits for the callback to return.
func (p *Page) runAfterRedraw() {
	fn := p.afterRedraw.fn
	if fn == nil || p.afterRedraw.busy {
		return
	}

	p.afterRedraw.busy = true
	defer func() { p.afterRedraw.busy = false }()

	fn()
}
