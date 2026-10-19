package csspreview

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

type demo struct {
	p          *ownframe.Page
	kind       string
	v          *view
	gen        uint64
	samples    map[string]*sample
	pendingKey string
}

type sample struct {
	base   []paint
	plan   motion
	origin time.Time
	active bool
}

func newDemo(p *ownframe.Page, kind string, v *view) *demo {
	return &demo{p: p, kind: kind, v: v, gen: p.Generation(), samples: map[string]*sample{}}
}

func (d *demo) tick(ctx context.Context) error {
	if len(d.samples) == 0 && d.pendingKey == "" {
		return nil
	}
	if d.pendingKey != "" {
		d.begin()
		d.rebind()
		d.gen = d.p.Generation()
		d.pendingKey = ""
	} else if d.gen != d.p.Generation() {
		d.rebind()
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

func (d *demo) rebind() {
	for key, s := range d.samples {
		if !s.active {
			continue
		}
		t, ok := targetByKey(d.kind, key)
		if !ok {
			continue
		}
		box, ok := findBox(d.p.Boxes(), t.id)
		if !ok {
			continue
		}
		s.plan = bind(d.p.Display(), box, t, s.base, s.origin)
	}
}
