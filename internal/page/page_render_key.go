package page

import (
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// renderTrack remembers the last committed frame so a Redraw that changes
// none of its inputs can return without laying the page out again.
type renderTrack struct {
	generation uint64
	imagesRev  uint64
	last       draw.Key
	hasLast    bool
	// statePseudos records the interactive pseudo-classes the sheets use. All
	// true means unknown, which keeps the old redraw behavior.
	statePseudos render.StatePseudos
	// redrawBatch defers draws across a pointer sequence. Any Redraw in the
	// batch sets pending; a Redraw (not RedrawIfChanged) also sets force.
	redrawBatch   int
	redrawPending bool
	redrawForce   bool
	// afterRedraw is the hook SetAfterRedraw registers, with the guards
	// that keep it from re-entering a redraw.
	afterRedraw draw.After
}

func (p *Page) viewport() draw.Viewport {
	return draw.Viewport{Width: p.width, Height: p.height}
}

func (p *Page) keyFor(state render.State) draw.Key {
	return draw.KeyFor(p.source, p.viewport(), state, p.statePseudos, p.imagesRev)
}

func (p *Page) unchanged(state render.State) bool {
	return draw.Unchanged(p.hasLast, p.last, p.keyFor(state))
}

func (p *Page) remember(state render.State) {
	p.last = p.keyFor(state)
	p.hasLast = true
}
