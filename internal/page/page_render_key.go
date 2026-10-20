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
}

func (p *Page) keyFor(state render.State) renderKey {
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
