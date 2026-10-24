package motion

import (
	"image"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// Step samples every bound player at now and returns the region the page
// must repaint: for each player, the union of its painted bounds before and
// after the sample. The set is fixed size, so a live frame allocates
// nothing.
//
// A sample a 2D matrix cannot carry leaves one Request per element, which
// TakeRestamp hands to the page's re-projection pass: a 3D transform, a
// perspective, or both on one element. A 2D-only page collects none.
//
// Step rebinds when the display it is handed is not the one the entries were
// resolved against. A page normally binds at every publish, so this guard
// only covers a list that reached Step without the hook; the original start
// survives the rebind, so a phase is not restarted and no player writes
// through an operation that moved.
func (r *Runtime) Step(d *Display, now time.Time) Damage {
	var damage Damage
	if r == nil {
		return damage
	}

	if d == nil {
		_ = r.Bind(nil)

		return damage
	}

	if d != r.display {
		// The bind failure is already kept in Stats; the tick does not
		// return it. A box that failed is simply not sampled.
		_ = r.Bind(d)
	}

	r.stats.Active, r.stats.Sampled, r.stats.NeedsRestamp = 0, 0, 0
	r.stats.Restamped = 0
	r.restamp = r.restamp[:0]
	for i := range r.entries {
		e := &r.entries[i]
		if e.mask == 0 {
			continue
		}

		before := e.bounds(d)
		active := e.player.Step(now)
		e.sampled = true
		r.stats.Sampled++
		if active {
			r.stats.Active++
		}
		if values := e.player.RestampValues(); len(values) > 0 {
			r.stats.NeedsRestamp++
			r.addRestamp(e, values)
		}

		after := before
		if e.mask.geometry() {
			after = e.bounds(d)
		}
		damage.Add(before.Union(after))
	}

	return damage
}

// bounds is the union of the painted boxes of the entry's target
// operations. Every index resolved at bind, so the range check only guards a
// list a caller shrank past its bind.
func (e *entry) bounds(d *Display) image.Rectangle {
	var box image.Rectangle

	for _, index := range e.indices {
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		box = box.Union(dirty.OpBounds(&d.Ops[index], d))
	}

	return box
}
