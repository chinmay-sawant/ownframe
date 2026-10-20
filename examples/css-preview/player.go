package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// partialLimit is the most samples whose damage still fits the window's
// eight-rect partial replay. Past it the union covers the page, and a full
// replay is cheaper than a tile rebuild.
const partialLimit = 8

type demo struct {
	p          *ownframe.Page
	v          *view
	gen        uint64
	samples    map[string]*sample
	pending    []string
	allPlaying bool
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

func (d *demo) tick(ctx context.Context) error {
	if len(d.samples) == 0 && len(d.pending) == 0 {
		d.p.UseFrameDirty()
		return nil
	}
	if len(d.pending) > 0 {
		pending := d.pending
		d.pending = nil
		for _, key := range pending {
			started, err := d.begin(key)
			if err != nil {
				return err
			}
			if !started {
				d.pending = append(d.pending, key)
			}
		}
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	} else if d.gen != d.p.Generation() {
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
	}
	active := 0
	for _, s := range d.samples {
		if s.active {
			active++
		}
	}
	partial := active > 0 && active <= partialLimit
	if partial || active == 0 {
		d.p.UseFrameDirty()
	}
	now := time.Now()
	live, waiting := d.stepSamples(now, partial)
	switch {
	case live:
		d.p.WantFrameIn(0)
	case waiting:
		d.p.WantFrameIn(50 * time.Millisecond)
	default:
		d.p.SetTick(nil)
	}
	return nil
}
