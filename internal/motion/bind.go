package motion

import (
	"errors"
	"fmt"
	"image"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/animation"
	"github.com/chinmay-sawant/ownframe/internal/cssanimation"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
)

// entry is one sampled player: the element it belongs to, the operations the
// player targets, the fields it may write, and the clock it keeps across
// rebinds. One element can carry several entries when its keyframes write
// different target sets, such as a transform on the subtree and a background
// colour on the element's own fill.
type entry struct {
	key     entryKey
	player  *animation.Player
	indices []int
	start   time.Time
	mask    fieldMask
	// pageX and pageY are the element's border-box origin in layout points. A
	// sampled perspective-origin is box-relative; the request needs the page
	// point the engine resolves.
	pageX, pageY float64
	// sampled is set by the first Step after a bind, so Active can ask for
	// one frame before the player has reported its own state.
	sampled bool
	// idle marks a player whose declared effects are all paused, triggered,
	// or unnamed. Sampling it cannot make it active, so a paused sample
	// never asks for a frame, not even the first one after a bind.
	idle bool
}

// entryKey matches an entry across a publish. Ordinal counts repeated ids,
// so a run of anonymous boxes keeps its phase too.
type entryKey struct {
	id      string
	ordinal int
}

// Bind resolves the animations declared by every box in d and builds the
// players each element needs, one per target set. A page calls it after it
// publishes a display list.
//
// An element that was bound before keeps its original start, so a rebind on
// a redraw, a patch, a hover, or a reload samples the same phase instead of
// restarting. The display a rebind captures is expected to carry the
// element's declared paint, as a fresh layout does; a display that still
// carries a sampled paint would compose from that sample.
//
// A box that fails to bind is left out and the failure is kept in Stats; the
// other boxes still animate. Bind(nil) releases every player.
//
// Bind also reports whether any bound player can need the engine's 3D
// re-projection pass (Stats.NeedsRetainedState): a 3D transform, a perspective,
// or a target element that already declares a 3D transform. The page turns
// retained layout state on before the first sample, because the pass cannot run
// without it.
func (r *Runtime) Bind(d *Display) error {
	if r == nil {
		return nil
	}

	if d == nil {
		r.display = nil
		r.entries = r.entries[:0]
		r.restamp = r.restamp[:0]
		r.stats.Bound, r.stats.Active, r.stats.Sampled = 0, 0, 0
		r.stats.NeedsRestamp = 0
		r.stats.NeedsRetainedState = false
		r.stats.Restamped, r.stats.RestampError = 0, ""
		r.stats.LastError = ""
		r.stats.Gaps = nil

		return nil
	}

	if len(r.entries) > 0 {
		r.stats.Rebinds++
	}

	starts := make(map[entryKey]time.Time, len(r.entries))
	for i := range r.entries {
		starts[r.entries[i].key] = r.entries[i].start
	}

	var ordinals map[string]int
	var failures []error
	var gaps []string
	needsRetained := false
	scale := displayPointsPerPixel(d)

	entries := r.entries[:0]
	for i := range d.Boxes {
		box := d.Boxes[i]
		if len(box.Animations) == 0 {
			continue
		}

		if ordinals == nil {
			ordinals = make(map[string]int)
		}
		ordinal := ordinals[box.ID]
		ordinals[box.ID] = ordinal + 1
		key := entryKey{id: box.ID, ordinal: ordinal}

		// The loop index is the box's position in d.Boxes, which is what
		// OpOwners points at. An element with no id is anonymous for owner
		// selection; BindBox then falls back to the subtree and reports it.
		boxIndex := -1
		if box.ID != "" {
			boxIndex = i
		}

		start, ok := starts[key]
		if !ok {
			start = r.clock()
		}

		bounds, boxGaps, err := cssanimation.BindBox(d, box, boxIndex, start)
		if err != nil {
			failures = append(failures, fmt.Errorf("motion: box %q: %w", box.ID, err))

			continue
		}
		gaps = append(gaps, boxGaps...)
		for j := range bounds {
			if gap := opacityGap(d, box, bounds[j]); gap != "" {
				gaps = append(gaps, gap)
			}
			needsRetained = needsRetained || bounds[j].Player.WantsRestamp()

			entries = append(entries, entry{
				key:     key,
				player:  bounds[j].Player,
				indices: bounds[j].Indices,
				start:   start,
				mask:    fieldMaskOf(bounds[j].Mask),
				pageX:   box.X * scale,
				pageY:   box.Y * scale,
				idle:    !runnable(box.Animations),
			})
		}
	}

	r.entries = entries
	r.display = d
	r.restamp = r.restamp[:0]
	r.stats.Bound = len(entries)
	r.stats.Active, r.stats.Sampled = 0, 0
	r.stats.NeedsRestamp = 0
	r.stats.NeedsRetainedState = needsRetained
	r.stats.Gaps = gaps

	err := errors.Join(failures...)
	if err != nil {
		r.stats.LastError = err.Error()
	} else {
		r.stats.LastError = ""
	}

	return err
}

// runnable reports whether any listed animation starts on its own. A paused
// sample waits for a class change and a triggered sample waits for an event;
// neither asks for a frame. An unnamed sample has no keyframes to apply.
func runnable(items []layout.Animation) bool {
	for i := range items {
		if items[i].PlayState == "paused" {
			continue
		}
		if items[i].Trigger != "" && items[i].Trigger != "none" {
			continue
		}
		if items[i].Name == "" || items[i].Name == "none" {
			continue
		}

		return true
	}

	return false
}

// opacityScanLimit bounds the pairwise overlap scan at bind. A subtree larger
// than this keeps the per-operation opacity path without claiming a gap, so one
// publish never pays a quadratic scan over a large animated card.
const opacityScanLimit = 128

// opacityGap reports a per-operation opacity that will not match a group alpha:
// the player fades two operations from separate branches of the element's
// subtree whose painted boxes overlap. Two operations in one ancestor chain
// stack opaquely, so per-operation alpha matches a group alpha there.
//
// Phase 08 replaces the per-operation write with a composited layer. Until
// then the gap is counted once per element, not written as if it were exact.
func opacityGap(d *Display, box layout.Box, bound cssanimation.Bound) string {
	if bound.Mask&cssmotion.MaskOpacity == 0 || len(bound.Indices) < 2 ||
		len(bound.Indices) > opacityScanLimit {
		return ""
	}
	for i := 0; i < len(bound.Indices); i++ {
		for j := i + 1; j < len(bound.Indices); j++ {
			a := &d.Ops[bound.Indices[i]]
			b := &d.Ops[bound.Indices[j]]
			if dirty.OpBounds(a, d).Intersect(dirty.OpBounds(b, d)) == (image.Rectangle{}) {
				continue
			}
			if !separateBranches(d, bound.Indices[i], bound.Indices[j]) {
				continue
			}
			return fmt.Sprintf(
				"box %q: opacity on overlapping operations is per-operation until phase 08", box.ID)
		}
	}
	return ""
}

// separateBranches reports whether two operations are painted by boxes that are
// neither the same box nor one an ancestor of the other, which is the overlap a
// group alpha would flatten and a per-operation alpha cannot.
func separateBranches(d *Display, first, second int) bool {
	if len(d.OpOwners) != len(d.Ops) {
		return false
	}
	a, b := d.OpOwners[first], d.OpOwners[second]
	if a < 0 || b < 0 || a == b {
		return false
	}
	return !ancestorOf(d, a, b) && !ancestorOf(d, b, a)
}

// ancestorOf reports whether ancestor is box or one of its box ancestors.
func ancestorOf(d *Display, ancestor, box int) bool {
	for depth := 0; depth <= len(d.Boxes) && box >= 0 && box < len(d.Boxes); depth++ {
		if box == ancestor {
			return true
		}
		box = d.Boxes[box].ParentIndex
	}
	return false
}

// displayPointsPerPixel returns the display's point-to-pixel factor, or 1 for a
// hand-built display that left it zero. A zero factor would collapse every page
// coordinate to the origin.
func displayPointsPerPixel(d *Display) float64 {
	if d.PointsPerPixel > 0 {
		return d.PointsPerPixel
	}

	return 1
}
