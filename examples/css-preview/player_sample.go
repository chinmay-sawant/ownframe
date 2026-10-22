package main

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

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
	w, h := p.Size()

	return &demo{p: p, v: v, gen: p.Generation(), samples: map[string]*sample{},
		treeParses: p.Stats().Parses, sizeW: w, sizeH: h}
}
