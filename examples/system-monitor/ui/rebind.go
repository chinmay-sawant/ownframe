package ui

import (
	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// accent is the single accent color, #2f6fd0. Every graph bar uses it, so
// the tick finds a card's bars by color inside the card's graph box.
var accent = [3]float64{0x2f / 255.0, 0x6f / 255.0, 0xd0 / 255.0}

// graphColsMax caps the bar nodes per graph. The tick lays them out at one
// bar per few visible pixels and downsamples history into at most this many
// columns, so paint work never follows the history size.
const graphColsMax = 96

// textMapCap pre-sizes the text handle map for the panels.
const textMapCap = 24

// handles are the retained operation pointers the tick paints. Redraw and
// resize replace the display list, so the tick reacquires whenever the
// page generation changes.
type handles struct {
	gen   uint64
	graph map[string]ownframe.Box
	bars  map[string][]*ownframe.DisplayOp
	text  map[string]*ownframe.DisplayOp
}

// rebind reacquires operation pointers from the current display list.
func (a *App) rebind() {
	s := a.state
	gen := a.page.Generation()
	s.lastGen = gen

	d := a.page.Display()
	if d == nil {
		// The page fell back to the bitmap path; there is nothing to
		// repaint in place.
		s.h = handles{gen: gen}

		return
	}

	h := handles{
		gen:   gen,
		graph: make(map[string]ownframe.Box, len(panelOrder)),
		bars:  make(map[string][]*ownframe.DisplayOp, len(panelOrder)),
		text:  make(map[string]*ownframe.DisplayOp, textMapCap),
	}
	boxes := a.page.Boxes()

	for _, p := range panelOrder {
		if box, ok := boxByID(boxes, p+"-graph"); ok {
			h.graph[p] = box
			h.bars[p] = frame.Fills(d, box, accent)
		}
	}

	for _, id := range paintIDs() {
		if box, ok := boxByID(boxes, id); ok {
			h.text[id] = frame.Text(d, box)
		}
	}

	s.h = h
}

// valid reports whether the retained handles still match the page.
func (a *App) valid() bool {
	return a.state.h.gen == a.page.Generation()
}
