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
	// playAllAt is the shared start time of a Play all run. Every sample
	// the run queues starts from it, so a row that scrolls in later shows
	// its animation at the phase the run reached. It is stamped after the
	// click has patched the cards, so the run starts when the click ends.
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
}

type sample struct {
	base   []paint
	to     []paint
	source string
	plan   motion
	origin time.Time
	active bool
	// detached marks a sample whose row left the window. It keeps its
	// origin so the animation resumes where it was when the row returns.
	detached bool
}

func newDemo(p *ownframe.Page, v *view) *demo {
	return &demo{p: p, v: v, gen: p.Generation(), samples: map[string]*sample{},
		treeParses: p.Stats().Parses}
}
