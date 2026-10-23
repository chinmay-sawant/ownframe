package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// invalidateCache drops the parsed tree so the next Redraw rebuilds it.
func (p *Page) invalidateCache() {
	p.cache = nil
	// Unknown sheets keep every state redraw until the next Redraw reports
	// what the new sheets can match.
	p.statePseudos = draw.UnknownPseudos()
}
