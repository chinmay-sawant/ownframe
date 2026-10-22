package main

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type demo struct {
	p          *ownframe.Page
	v          *view
	gen        uint64
	samples    map[string]*sample
	pending    []string
	allPlaying bool
	// pendingAt is when the pending queue was last drained, so a row that
	// cannot start yet does not cost a box scan on every frame.
	pendingAt time.Time
	// playAllAt is when a Play all run started. A run that has not begun
	// yet carries no stamp.
	playAllAt  time.Time
	startedAt  time.Time
	startedSeq uint64
	// prevSelected is the card that loses the selection when a new card is
	// selected.
	prevSelected string
	// readout is the text the readout shows. It lives here and not in the
	// template data, because the readout is patched into the parsed tree
	// and a template that carried it would rebuild the page on every
	// click.
	readout string
	// treeDirty marks that the parsed tree carries state the display has
	// not shown yet.
	treeDirty bool
	// dataChanged marks that the click changed the template data, so the
	// page has to draw again.
	dataChanged bool
	// treeParses is the parse count when the app last read the tree.
	treeParses uint64
	// window is the row window that slices the catalog.
	window *catalogWindow
	// sizeW and sizeH are the page size the geometry was measured at.
	sizeW, sizeH int
	// remeasure marks that the window size changed, so the catalog has to
	// be measured again.
	remeasure bool
	// resizeAt is when the window size last changed.
	resizeAt time.Time
}
