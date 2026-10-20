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
	// its animation at the phase the run reached.
	playAllAt  time.Time
	startedAt  time.Time
	startedSeq uint64
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
	return &demo{p: p, v: v, gen: p.Generation(), samples: map[string]*sample{}}
}
