package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
)

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

	p.afterRedraw.Set(fn)
}

// runAfterRedraw calls the registered callback after a commit installed the
// display and the boxes. A callback that asks for a redraw does not re-enter
// the commit: the request waits for the callback to return.
func (p *Page) runAfterRedraw() {
	p.afterRedraw.Run()
}

// afterRedrawDraw serves the redraw the callback asked for. The draw runs
// after the callback returned, so the commit is never re-entered, and the
// callback runs again in it, so state reapplied to a rebuilt tree lands. One
// draw is enough: a callback that always asks for a draw cannot loop.
func (p *Page) afterRedrawDraw(ctx context.Context) error {
	if !p.afterRedraw.Take() {
		return nil
	}

	defer p.afterRedraw.Done()

	return p.redraw(ctx, false)
}

// BeginRedrawBatch defers page draws until EndRedrawBatch. The window wraps a
// press, focus, click, and release in one batch so the final state draws once.
//
// A redraw that a failed draw left queued stays queued, so the flags survive
// into the next batch and it commits the redraw.
func (p *Page) BeginRedrawBatch() {
	draw.OpenBatch(&p.redrawBatch)
}

// EndRedrawBatch closes a batch and draws the final state once. An inner batch
// leaves the draw to the outermost one. A plain Redraw in the batch forces the
// final draw; only RedrawIfChanged calls let the final draw skip unchanged
// state.
//
// A draw that fails leaves the redraw queued, so the next batch commits it
// and the committed frame catches up with the state the batch changed.
func (p *Page) EndRedrawBatch(ctx context.Context) error {
	run, force := draw.CloseBatch(&p.redrawBatch, &p.redrawPending, &p.redrawForce)
	if !run {
		return nil
	}

	if err := p.redraw(ctx, !force); err != nil {
		draw.RestoreBatch(&p.redrawPending, &p.redrawForce, force)

		return err
	}

	return nil
}
