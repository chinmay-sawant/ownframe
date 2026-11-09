package page

import "github.com/chinmay-sawant/ownframe/internal/page/draw"

// invalidateCache drops the parsed tree so the next Redraw rebuilds it. It
// also evicts the box-cache entry for the discarded boxes, so a source
// change on a live page does not pin the old box list until the next commit.
func (p *Page) invalidateCache() {
	p.cache = nil
	p.relayedOK = false
	p.forgetBoxes()
	// Unknown sheets keep every state redraw until the next Redraw reports
	// what the new sheets can match.
	p.statePseudos = draw.UnknownPseudos()
}
