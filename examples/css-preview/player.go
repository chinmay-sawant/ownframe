package main

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type demo struct {
	p          *ownframe.Page
	v          *view
	gen        uint64
	samples    map[string]*sample
	pendingKey string
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
	active := false
	for _, s := range d.samples {
		if s.active && !s.plan.step(time.Now()) {
			s.active = false
		}
		active = s.active || active
	}
	if !active {
		d.p.SetTick(nil)
	}
	return nil
}

func (d *demo) rebind() error {
	for key, s := range d.samples {
		if !s.active {
			continue
		}
		t, ok := targetByKey(key)
		if !ok {
			continue
		}
		boxes := d.p.Boxes()
		boxIndex, ok := findBoxIndex(boxes, t.id)
		if !ok {
			continue
		}
		box := boxes[boxIndex]
		plan, err := bind(d.p.Display(), box, t, s.indices, s.base, s.origin)
		if err != nil {
			return err
		}
		s.plan = plan
		if key == d.pendingKey && plan.animation != nil {
			if err := dispatchAnimationTriggers(boxes, boxIndex, s.source, plan.animation, s.origin); err != nil {
				return err
			}
		}
	}
	return nil
}
