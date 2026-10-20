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
	pendingKey string
	startedAt  time.Time
	startedSeq uint64
}

type sample struct {
	base    []paint
	indices []int
	source  string
	plan    motion
	origin  time.Time
	active  bool
}

func newDemo(p *ownframe.Page, v *view) *demo {
	return &demo{p: p, v: v, gen: p.Generation(), samples: map[string]*sample{}}
}

func (d *demo) tick(ctx context.Context) error {
	if len(d.samples) == 0 && d.pendingKey == "" {
		return nil
	}
	if d.pendingKey != "" {
		if err := d.begin(); err != nil {
			return err
		}
		if err := d.rebind(); err != nil {
			return err
		}
		d.gen = d.p.Generation()
		d.pendingKey = ""
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
	if partial {
		d.p.UseFrameDirty()
	}
	now := time.Now()
	live := false
	for _, s := range d.samples {
		if !s.active {
			continue
		}
		var running bool
		if partial {
			running = d.stepSample(s, now)
		} else {
			running = s.plan.step(now)
		}
		if !running {
			s.active = false
		}
		live = s.active || live
	}
	if live {
		d.p.WantFrameIn(0)
	} else {
		d.p.SetTick(nil)
	}
	return nil
}
