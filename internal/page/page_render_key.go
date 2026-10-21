package page

import (
	"github.com/chinmay-sawant/blinkless/css"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// renderKey is every input that changes the layout of a frame. Images are
// represented by a counter because a resolver function cannot be compared.
type renderKey struct {
	source               string
	width, height        int
	focus, hover, active string
	theme                *css.Sheet
	images               uint64
}

// renderTrack remembers the last committed frame so a Redraw that changes
// none of its inputs can return without laying the page out again.
type renderTrack struct {
	generation uint64
	imagesRev  uint64
	last       renderKey
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
	afterRedraw afterRedrawState
}

func (p *Page) keyFor(state render.State) renderKey {
	// State the sheets cannot match never changes paint, so leaving it out of
	// the key lets a hover or press land without forcing a layout.
	if !p.statePseudos.Hover {
		state.Hover = ""
	}

	if !p.statePseudos.Active {
		state.Active = ""
	}

	if !p.statePseudos.Focus {
		state.Focus = ""
	}

	return renderKey{
		source: p.source, width: p.width, height: p.height,
		focus: state.Focus, hover: state.Hover, active: state.Active,
		theme: state.Theme, images: p.imagesRev,
	}
}

func (p *Page) unchanged(state render.State) bool {
	return p.hasLast && p.last == p.keyFor(state)
}

func (p *Page) remember(state render.State) {
	p.last = p.keyFor(state)
	p.hasLast = true
}
