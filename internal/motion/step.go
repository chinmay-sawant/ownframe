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
		r.ReleaseTransitions()

		return damage
	}

	if d != r.display {
		// The bind failure is already kept in Stats; the tick does not
		// return it. A box that failed is simply not sampled.
		_ = r.Bind(d)
	}

	r.stats.Active, r.stats.Sampled, r.stats.NeedsRestamp = 0, 0, 0
	r.stats.Restamped = 0
	r.stats.InactiveTimelines = 0
	r.stats.TransitionsActive, r.stats.TransitionsSampled = 0, 0
	r.stats.SVGPaintRequests, r.stats.SVGPaintRestyles = 0, 0
	r.stats.SVGPaintQueued, r.stats.SVGPaintBudgetHit = 0, false
	r.stats.SVGPaintDeferred, r.stats.SVGPaintDeferredReason = 0, ""
	r.restamp = r.restamp[:0]
	r.svg.requests = r.svg.requests[:0]
	damage = r.stepTransitions(d, now)
	for i := range r.entries {
		e := &r.entries[i]
		if e.layout || (e.mask == 0 && !e.svg) {
			// The layout player owns no paint target; StepLayout samples it
			// before this pass, against the display the relayout produced.
			continue
		}

		// A paused player is sampled once, to reach the pose it froze at, and
		// then skipped: repeating the same sample would dirty the page every
		// tick and keep a paused window painting. An SVG entry is the
		// exception: its held sample may still be waiting for the one-frame
		// retained-state redraw, and collecting it again is what keeps the
		// request alive until the page can apply it.
		paused := e.player.Paused()
		if paused && e.wasPaused && e.sampled && !e.svg {
			continue
		}
		e.wasPaused = paused

		before := e.bounds(d)
		input, inactive := r.input(e, now)
		if inactive > 0 {
			r.stats.InactiveTimelines++
		}
		active := e.player.StepWithInput(input)
		e.sampled, e.sampledAt = true, now
		r.stats.Sampled++
		if active {
			r.stats.Active++
		}
		if values := e.player.RestampValues(); len(values) > 0 {
			r.stats.NeedsRestamp++
			r.addRestamp(e, values)
		}
		if e.svg {
			r.collectSVGPaint(e)
		}

		after := before
		if e.mask.geometry() {
			after = e.bounds(d)
		}
		// An SVG entry writes no operation; the page marks the engine patch's
		// image bounds when it adopts the restyle. Marking the unchanged image
		// box here would repaint a raster that has not been rebuilt yet.
		// An offscreen entry still samples, so its phase is right when it
		// scrolls in, but its write needs no repaint. A promoted entry writes
		// into its retained layer, so the page background under it does not
		// change either. A player that did no work holds the pose it already
		// presented, so only its first resting sample can still change the
		// screen; a repeated rest pose marks nothing and a settled page
		// parks instead of repainting static elements every tick.
		if !e.svg && r.visible(e) && !r.promotedEntry(e) && freshPaint(e, active) {
			damage.Add(before.Union(after))
		}
	}

	return damage
}

// freshPaint reports that this sample may have changed the element's paint,
// so the frame must cover its bounds. A player that did no work holds the
// pose it already presented; only its first resting sample can still change
// the screen (a fill mode snapping back to base, or a rebind repainting from
// base), so a repeated rest pose marks nothing and a settled page parks.
func freshPaint(e *entry, active bool) bool {
	if active {
		e.settled = false
		return true
	}
	if e.settled {
		return false
	}
	e.settled = true
	return true
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
