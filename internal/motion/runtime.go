// Package motion drives the CSS keyframe animations a page declares. A page
// binds the runtime after it publishes a display list and steps it once per
// frame; the runtime resolves one player per animated element, samples the
// players, and reports the rectangles the page must repaint.
//
// The package does not import internal/page, internal/window, or
// internal/replay. The page owns the display list, and motion only edits the
// operations it was handed.
package motion

import (
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
)

// Display is the retained display list a runtime samples. It is the same
// layout type the page publishes, so a caller passes its list unchanged.
type Display = layout.Display

// Runtime binds the keyframe animations of one page and samples them.
// It is not safe for concurrent use: the page loop owns it.
type Runtime struct {
	// display is the list the entries were resolved against. A publish
	// installs a new list pointer, so a mismatch in Step means the page
	// replaced the list without rebinding.
	display *Display

	entries []entry
	stats   Stats

	// transitions is the second registry: one renderer per element, property,
	// and target set. It shares the runtime and the Step call with the
	// animation players, but it is bound at a display commit rather than at a
	// stylesheet bind.
	transitions []transitionEntry

	// geometry is the layout-length registry. It shares the relayout queue and
	// budget with the layout keyframe driver, so a width or height transition
	// reaches the engine's override path once per frame.
	geometry []geometryEntry

	// discrete is the non-interpolated registry: presence and stacking rules
	// that install one value at a progress boundary instead of blending.
	discrete []discreteEntry

	// transitionCapture is the capture-side diagnostic block the next bind
	// copies into Stats.
	transitionCapture transitionCaptureCounts

	// view is the page scroller state a scroll or view timeline samples
	// against. hasView separates an unscrolled page at offset zero from a
	// runtime no page has configured.
	view    View
	hasView bool

	// restamp is the per-element re-projection requests the last Step
	// collected. TakeRestamp hands them to the page and clears them, so one
	// tick applies one pass.
	restamp []Request

	// layout is the layout-tier driver: the per-frame relayout budget, the
	// element queue, and the sampled requests. StepLayout fills it and the
	// page applies the engine patch.
	layout layoutState

	// svg is the inline-SVG paint driver: the paint already applied per
	// element, the requests the latest Step collected, and the per-frame
	// restyle budget. The page applies the requests through
	// layout.RestyleInlineSVG.
	svg svgState

	// layerEntries is the LayerEntries scratch, reused so a per-frame
	// promotion read allocates nothing.
	layerEntries []LayerEntry

	// now is the clock a newly bound effect starts at. Tests replace it.
	now func() time.Time
}

// New returns an idle runtime with no players bound.
func New() *Runtime {
	return &Runtime{now: time.Now}
}

// Active reports whether the runtime needs another frame: a time-driven
// player has not been sampled yet, or its latest sample left work to do. A
// player made only of scroll, view, or named timeline effects is not a reason
// to tick: its pose changes when the page scrolls, and a still scroll position
// is a static sample. A finished, paused, or triggered player is not active,
// so the page stops asking for frames.
func (r *Runtime) Active() bool {
	if r == nil {
		return false
	}
	if r.animationsActive() {
		return true
	}
	for i := range r.transitions {
		e := &r.transitions[i]
		if !e.sampled || e.active {
			return true
		}
	}
	if r.geometryActive() || r.discreteActive() {
		return true
	}
	return false
}

// animationsActive reports whether any keyframe player needs a frame.
func (r *Runtime) animationsActive() bool {
	for i := range r.entries {
		e := &r.entries[i]
		if !e.timed {
			continue
		}
		if e.layout {
			// A layout player is sampled by StepLayout, not Step, so its
			// queue membership is its work signal.
			if r.layout.has(e.key.id) {
				return true
			}

			continue
		}
		if !e.sampled && !e.idle {
			return true
		}

		if e.player.Active() {
			return true
		}
	}

	return false
}

// NextFrame returns the next time the runtime needs a sample. An animation and
// a running transition ask for the next frame; a transition still inside its
// delay asks for the delay end instead, so a delayed page parks rather than
// repainting the same starting paint every vsync. ok is false when nothing is
// running.
func (r *Runtime) NextFrame(now time.Time) (time.Time, bool) {
	if r == nil {
		return time.Time{}, false
	}
	var wake time.Time
	found := false
	add := func(at time.Time) {
		if !found || at.Before(wake) {
			wake, found = at, true
		}
	}
	if r.animationsActive() {
		add(now)
	}
	for i := range r.transitions {
		e := &r.transitions[i]
		if !e.sampled {
			add(now)
			continue
		}
		if !e.active {
			continue
		}
		at := now
		if end := transitionDelayEnd(e); now.Before(end) {
			at = end
		}
		add(at)
	}
	if at, ok := r.nextGeometryFrame(now); ok {
		add(at)
	}
	if at, ok := r.nextDiscreteFrame(now); ok {
		add(at)
	}
	return wake, found
}

// Stats returns a snapshot of the runtime counters.
func (r *Runtime) Stats() Stats {
	if r == nil {
		return Stats{}
	}

	return r.stats
}

// Wake marks every sampled player as due for another sample. A trigger event
// changes a player's pose without writing the operations, and a paused player
// is normally sampled once and then skipped, so the page calls Wake after a
// dispatch and the next Step writes the trigger's pose. It does not allocate
// and it is safe between samples.
func (r *Runtime) Wake() {
	if r == nil {
		return
	}

	for i := range r.entries {
		r.entries[i].wasPaused = false
		if r.entries[i].layout {
			r.entries[i].sampled = false
			r.layout.enqueue(r.entries[i].key.id)
		}
	}
}

// clock returns the time a newly bound effect starts at.
func (r *Runtime) clock() time.Time {
	if r.now == nil {
		return time.Now()
	}

	return r.now()
}

// TriggerPass is one player set whose animation records are aligned to that
// player's effect indices. csstrigger.Dispatch addresses an action by box
// index and animation index, so a page has to hand it a box list where each
// triggered element's animation list matches the one player in Players. One
// element can bind several players with disjoint target sets, so a dispatch
// runs once per pass and every pass is self-consistent.
type TriggerPass struct {
	// Boxes is a shallow copy of the display's box list with only this pass's
	// triggered animation records. Scope and parent links are unchanged.
	Boxes []layout.Box
	// Players maps a box index in Boxes to the player its records were
	// aligned to.
	Players map[int]*animation.Player
}

// TriggerPasses builds the dispatch views for the current display. It returns
// nil when no box declares an animation-trigger, so a page without triggers
// never pays for a pass. It allocates, so it is for event dispatch, not for
// the sample loop.
func (r *Runtime) TriggerPasses() []TriggerPass {
	if r == nil || r.display == nil || len(r.entries) == 0 {
		return nil
	}

	byBox := make(map[int][]*entry, len(r.entries))
	for i := range r.entries {
		e := &r.entries[i]
		if e.triggerAnims == nil {
			continue
		}
		byBox[e.boxIndex] = append(byBox[e.boxIndex], e)
	}
	if len(byBox) == 0 {
		return nil
	}

	passes := 0
	for _, group := range byBox {
		passes = max(passes, len(group))
	}

	out := make([]TriggerPass, 0, passes)
	for group := 0; group < passes; group++ {
		boxes := make([]layout.Box, len(r.display.Boxes))
		copy(boxes, r.display.Boxes)
		for i := range boxes {
			boxes[i].Animations = nil
		}
		players := make(map[int]*animation.Player, len(byBox))
		for boxIndex, entries := range byBox {
			if group >= len(entries) {
				continue
			}
			boxes[boxIndex].Animations = entries[group].triggerAnims
			players[boxIndex] = entries[group].player
		}
		out = append(out, TriggerPass{Boxes: boxes, Players: players})
	}

	return out
}
