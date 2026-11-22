package motion

import (
	"errors"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
	"slices"
	"sort"
	"strings"
	"time"
)

// entry is one sampled player: the element it belongs to, the operations the
// player targets, the fields it may write, and the clock it keeps across
// rebinds. One element can carry several entries when its keyframes write
// different target sets, such as a transform on the subtree and a background
// colour on the element's own fill.
type entry struct {
	key     entryKey
	player  *Player
	indices []int
	start   time.Time
	mask    fieldMask
	// paintMask is the CSS paint mask the player may write. It narrows the
	// paused pose a rebind copies into the new operations, so a transform
	// player never overwrites a sibling colour player's values.
	paintMask Mask
	// pageX and pageY are the element's border-box origin in layout points. A
	// sampled perspective-origin is box-relative; the request needs the page
	// point the engine resolves.
	pageX, pageY float64
	// boxX, boxY, boxW, and boxH are the element's border box in CSS pixels,
	// page space. Scroll progress, view progress, and the named range values
	// are resolved against it.
	boxX, boxY, boxW, boxH float64
	// boxIndex is the element's position in the display box list. Trigger
	// dispatch addresses its player by that index.
	boxIndex int
	// timelines is the resolved animation-timeline state for this element.
	timelines timelineState
	// timed reports that at least one named animation follows the wall clock,
	// so the player may ask for another frame. A player made only of scroll,
	// view, or named timeline effects changes only when the page scrolls.
	timed bool
	// triggerAnims is this player's animation records aligned to its effect
	// indices, with the trigger string each effect was built from. It is nil
	// when the element declares no animation-trigger.
	triggerAnims []layout.Animation
	// sampled is set by the first Step after a bind, so Active can ask for
	// one frame before the player has reported its own state.
	sampled bool
	// keptPose means the new display still carries the previous player's
	// sample. Another bind before Step must carry its original base again.
	keptPose bool
	// sampledAt is the clock of the latest Step. A rebind that pauses an
	// animation freezes at this time, so the pose on screen is the pose the
	// paused player shows rather than one advanced by the redraw.
	sampledAt time.Time
	// wasPaused records the play state of the latest Step. A paused player is
	// sampled once to reach its pose and then skipped, so a page that is
	// paused parks instead of rewriting the same paint every tick.
	wasPaused bool
	// settled records that the latest Step found the player at rest and
	// already marked that pose. A repeated resting sample paints nothing
	// new, so Step marks no damage for it and a settled page parks; the
	// first resting sample still marks, because a fill mode can snap back
	// to base and a rebind repaints from base.
	settled bool
	// idle marks a player whose declared effects are all triggered or
	// unnamed. Sampling it cannot make it active, so it never asks for a
	// frame, not even the first one after a bind.
	idle bool
	// layout marks the player that samples the element's layout-tier
	// keyframes. Step skips it (it owns no paint target) and StepLayout
	// samples it under the relayout budget.
	layout bool
	// svg marks the player that samples the element's inline-SVG paint. It
	// targets the SVG root's image operation but writes no operation field, so
	// Step still samples it and the page restyles the raster afterwards.
	svg bool
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
// An element that was bound before keeps its clock, so a rebind on a redraw, a
// patch, a hover, or a reload samples the same phase instead of restarting. A
// player that was paused and still is keeps its frozen pose; one that changes
// play state pauses or resumes at the sample the paint came from, so the
// pause button neither snaps the element to its base pose nor restarts it. The
// display a rebind captures is expected to carry the element's declared paint,
// as a fresh layout does; a display that still carries a sampled paint would
// compose from that sample.
//
// A box that fails to bind is left out and the failure is kept in Stats; the
// other boxes still animate. Bind(nil) releases every player.
//
// Bind also reports whether any bound player can need the engine's retained
// layout state (Stats.NeedsRetainedState): a 3D transform, a perspective, a
// target element that already declares a 3D transform, or a layout-tier
// keyframe the override relayout path patches. The page turns retained layout
// state on before the first sample, because neither pass can run without it.
func (r *Runtime) Bind(d *Display) error {
	if r == nil {
		return nil
	}

	return r.BindAfter(r.display, d)
}

// AdoptDisplay installs next without rebuilding the players. A restamp mutates
// the same operation slice in place, so every target pointer and index stays
// valid; only the display pointer the runtime checks for a rebind changes.
func (r *Runtime) AdoptDisplay(next *Display) {
	if r == nil {
		return
	}

	r.display = next
	if r.index != nil {
		r.index.Repoint(next)
	}
}

// BindAfter resolves the animations of next, handing each player the clock and
// paused pose it had in previous. A page passes the display it drew last as
// previous, so a redraw that changes animation-play-state keeps the phase
// instead of starting over or snapping to the CSS base. Bind is BindAfter with
// the runtime's own last display.
//
// BindAfter(_, nil) releases every player.
func (r *Runtime) BindAfter(previous, next *Display) error {
	return r.BindAfterWithIndex(previous, next, nil)
}

// BindAfterWithIndex resolves the animations of next using the provided index.
func (r *Runtime) BindAfterWithIndex(previous, next *Display, index *Index) error {
	if r == nil {
		return nil
	}

	if next == nil {
		r.display = nil
		r.index = nil
		r.entries = r.entries[:0]
		r.restamp = r.restamp[:0]
		r.stats.Bound, r.stats.Active, r.stats.Sampled = 0, 0, 0
		r.stats.NeedsRestamp = 0
		r.stats.NeedsRetainedState = false
		r.stats.Restamped, r.stats.RestampError = 0, ""
		r.stats.LastError = ""
		r.stats.Gaps = nil
		r.stats.UnresolvedTimelines = 0
		r.stats.InactiveTimelines = 0
		r.stats.LayoutBound = 0
		r.stats.LayoutRelayouts = 0
		r.stats.LayoutQueued = 0
		r.stats.LayoutBudgetHit = false
		r.stats.LayoutError = ""
		r.layout.reset()
		r.svg.reset()
		r.stats.SVGPaintRequests = 0
		r.stats.SVGPaintRestyles = 0
		r.stats.SVGPaintQueued = 0
		r.stats.SVGPaintBudgetHit = false
		r.stats.SVGPaintDeferred = 0
		r.stats.SVGPaintDeferredReason = ""
		r.stats.SVGPaintError = ""

		return nil
	}

	if len(r.entries) > 0 {
		r.stats.Rebinds++
	}

	hasAnims := false
	for i := range next.Boxes {
		if len(next.Boxes[i].Animations) > 0 {
			hasAnims = true
			break
		}
	}
	if !hasAnims && len(r.entries) == 0 {
		r.display = next
		r.stats.Bound, r.stats.Active, r.stats.Sampled = 0, 0, 0
		r.stats.NeedsRestamp = 0
		r.stats.NeedsRetainedState = false
		r.stats.LastError = ""
		r.stats.Gaps = nil
		r.stats.UnresolvedTimelines = 0
		r.stats.InactiveTimelines = 0
		r.stats.LayoutBound = 0
		r.stats.SVGPaintRequests = 0
		return nil
	}

	// A publish rebuilds the display from the declared styles, so the paint a
	// previous tick restyled is gone. Clearing the applied signatures makes the
	// next Step request it again. A page that just adopted an SVG patch calls
	// MarkSVGPaintKept before this bind, because that display carries the
	// restyle and NoteSVGPaint records the paint it installed.
	if r.svg.keepBind {
		r.svg.keepBind = false
	} else {
		r.svg.forget()
	}

	now := r.clock()
	carried := r.snapshotCarry(previous)

	starts := make(map[entryKey]time.Time, len(r.entries))
	for i := range r.entries {
		starts[r.entries[i].key] = r.entries[i].start
	}

	var ordinals map[string]int
	var failures []error
	var gaps []string
	needsRetained := false
	unresolvedTimelines := 0
	scale := displayPointsPerPixel(next)
	// The operation index is built once for the display every selection
	// resolves against, so a bind with dozens of animated boxes shares one
	// ownership scan instead of repeating it per box per paint class. It is
	// built lazily: a page with no animations pays no scan.
	if index != nil && index != EmptyIndex() && index.Display() != next {
		if index.NumOps() == len(next.Ops) {
			index.Repoint(next)
		} else {
			index = nil
		}
	}

	// The old entries are no longer read: snapshotCarry and starts copied what
	// the replacement needs before the first append.
	entries := r.entries[:0]
	used := make(map[entryKey]int, len(carried))
	for i := range next.Boxes {
		box := next.Boxes[i]
		if len(box.Animations) == 0 {
			continue
		}

		if ordinals == nil {
			ordinals = make(map[string]int)
		}
		ordinal := ordinals[box.ID]
		ordinals[box.ID] = ordinal + 1
		key := entryKey{id: box.ID, ordinal: ordinal}

		// The loop index is the box's position in next.Boxes, which is what
		// OpOwners points at. An element with no id is anonymous for owner
		// selection; BindBox then falls back to the subtree and reports it.
		boxIndex := -1
		if box.ID != "" {
			boxIndex = i
		}

		start, ok := starts[key]
		if !ok {
			start = now
		}

		if index == nil || (index == EmptyIndex() && len(next.Keyframes) > 0) {
			index = r.indexFor(next)
		}
		bounds, boxGaps, err := BindBoxWithIndex(next, box, boxIndex, start, index)
		if err != nil {
			failures = append(failures, fmt.Errorf("motion: box %q: %w", box.ID, err))

			continue
		}
		gaps = append(gaps, boxGaps...)
		// The layout player is separate from the paint players: it owns no
		// operation, its keyframes are the layout tier, and StepLayout samples
		// it under the per-frame relayout budget.
		layoutPlayer, err := BindLayout(next, box, start)
		if err != nil {
			failures = append(failures, fmt.Errorf("motion: box %q: %w", box.ID, err))

			continue
		}
		timelines, unresolved := bindTimelines(next, box, i)
		unresolvedTimelines += unresolved
		triggerAnims, triggerGap := triggerAlignment(next, box, boxIndex, bounds, start, index)
		if triggerGap != "" {
			gaps = append(gaps, triggerGap)
		}
		timed := hasTimeEffect(box.Animations)
		sequence := carried[key]
		for j := range bounds {
			if gap := opacityGap(next, box, bounds[j]); gap != "" {
				gaps = append(gaps, gap)
			}
			needsRetained = needsRetained || bounds[j].Player.WantsRestamp()

			e := entry{
				key:       key,
				player:    bounds[j].Player,
				indices:   bounds[j].Indices,
				start:     start,
				mask:      fieldMaskOf(bounds[j].Mask),
				paintMask: bounds[j].Mask,
				pageX:     box.X * scale,
				pageY:     box.Y * scale,
				boxX:      box.X,
				boxY:      box.Y,
				boxW:      box.W,
				boxH:      box.H,
				boxIndex:  i,
				timelines: timelines,
				timed:     timed,
				idle:      !runnable(box.Animations),
				svg:       bounds[j].Mask&SVGPaintMask != 0,
			}
			if j < len(triggerAnims) {
				e.triggerAnims = triggerAnims[j]
			}
			if n := used[key]; n < len(sequence) && sequence[n].ok {
				restoreCarry(&e, sequence[n], next, now)
			}
			used[key]++
			entries = append(entries, e)
		}

		if layoutPlayer == nil {
			continue
		}
		if box.ID == "" {
			gaps = append(gaps,
				fmt.Sprintf("box %q: layout keyframes need an element id", box.ID))

			continue
		}
		needsRetained = true
		e := entry{
			key:       key,
			player:    layoutPlayer,
			start:     start,
			pageX:     box.X * scale,
			pageY:     box.Y * scale,
			boxX:      box.X,
			boxY:      box.Y,
			boxW:      box.W,
			boxH:      box.H,
			boxIndex:  i,
			timelines: timelines,
			timed:     timed,
			idle:      !runnable(box.Animations),
			layout:    true,
		}
		if n := used[key]; n < len(sequence) && sequence[n].ok {
			restoreCarry(&e, sequence[n], next, now)
		}
		used[key]++
		entries = append(entries, e)
	}

	r.entries = entries
	r.display = next
	r.index = index
	stampTransformClips(next, entries)
	r.restamp = r.restamp[:0]
	r.stats.Bound = len(entries)
	r.stats.Active, r.stats.Sampled = 0, 0
	r.stats.NeedsRestamp = 0
	r.stats.NeedsRetainedState = needsRetained
	r.stats.Gaps = gaps
	r.stats.UnresolvedTimelines = unresolvedTimelines
	r.stats.InactiveTimelines = 0
	r.syncLayoutQueue()

	err := errors.Join(failures...)
	if err != nil {
		r.stats.LastError = err.Error()
	} else {
		r.stats.LastError = ""
	}

	return err
}

// carry is one entry's handoff across a rebind: the clock its player reached,
// the time of its last sample, and the paused pose. A paused player samples
// once after the bind, but the frame drawn between the publish and that tick
// must already show the frozen pose.
type carry struct {
	clock     Clock
	sampledAt time.Time
	paints    []Paint
	// bases is the paint the previous player composed from, one per target.
	// A display that still carries that player's sample is repaired from it,
	// so the next sample composes from the declared pose.
	bases []Paint
	ok    bool
	// sampled reports that the entry was sampled before the rebind. The
	// layout queue uses it so a settled layout animation is not reintroduced
	// on every patch-triggered rebind.
	sampled bool
}

// snapshotCarry groups the handoff for every bound entry by element key. The
// per-key order is the order BindBox produced the target groups, so two
// players on one element (a transform group and a colour group, say) keep their
// own clocks and poses.
func (r *Runtime) snapshotCarry(previous *Display) map[entryKey][]carry {
	if len(r.entries) == 0 {
		return nil
	}
	source := previous
	if source == nil {
		source = r.display
	}
	groups := make(map[entryKey][]carry, len(r.entries))
	for i := range r.entries {
		e := &r.entries[i]
		var c carry
		if clock, ok := e.player.Clock(); ok {
			c.ok, c.clock, c.sampledAt = true, clock, e.sampledAt
			c.sampled = e.sampled
			if e.sampled || e.keptPose || clock.Paused {
				c.paints = pausedPaints(source, e)
				c.bases = e.player.BasePaints(c.bases)
			}
		}
		groups[e.key] = append(groups[e.key], c)
	}

	return groups
}

// pausedPaints reads the pose on screen for one entry, narrowed to the fields
// the player owns. A transform player copies only the matrix and a colour
// player only the channels, so the two entries sharing one operation do not
// overwrite each other on the way in.
func pausedPaints(source *Display, e *entry) []Paint {
	if source == nil {
		return nil
	}
	paints := make([]Paint, len(e.indices))
	for k, index := range e.indices {
		if index < 0 || index >= len(source.Ops) {
			// The entries were resolved against this list, so a miss means
			// the caller handed a different display; copying part of a pose
			// would be worse than showing the declared paint.
			return nil
		}
		paint := Read(&source.Ops[index])
		paint.Mask = e.paintMask
		paints[k] = paint
	}

	return paints
}

// restoreCarry puts a rebuilt player on the clock its predecessor carried. A
// player that was paused and still is keeps its frozen offset. A paused player
// that is running again starts a new run at now, so the paused time is not
// counted. A running player that the new stylesheet paused freezes at its last
// sample, which is the pose already on screen. In that last case, and when the
// element stays paused, the old pose is copied onto the new operations before
// the first tick so the redraw does not snap to the CSS base.
//
// The base paint is repaired first, because the display a rebind is handed is
// not always rebuilt from the declared styles: a re-projected patch is built
// from the retained layout result the previous sample wrote through, so its
// operations carry that sample. Composing the next sample from that pose would
// stack it on itself, which is the paused negative-delay drift.
func restoreCarry(e *entry, c carry, next *Display, now time.Time) {
	paused := e.player.Paused()
	restoreBases(e, c, next)
	if e.keptPose {
		e.sampledAt = c.sampledAt
	}
	clock := c.clock
	switch {
	case clock.Paused && paused:
		// The new player starts paused; its carried offset is the phase it
		// already had, so the clock needs no change.
	case clock.Paused && !paused:
		clock.Start = now
		clock.Paused, clock.PausedAt = false, time.Time{}
		clock.Position, clock.HasPosition = 0, false
	case !clock.Paused && paused:
		freeze := c.sampledAt
		if freeze.IsZero() || freeze.Before(clock.Start) || freeze.After(now) {
			freeze = now
		}
		clock.Offset += time.Duration(float64(freeze.Sub(clock.Start)))
		clock.Start, clock.Paused, clock.PausedAt = freeze, true, freeze
		clock.Position, clock.HasPosition = 0, false
	}
	e.player.SetClock(clock)
	e.start = clock.Start
	if e.layout {
		// The layout player samples only through the relayout queue, so it
		// keeps its own sampled flag across the rebind. A paint player keeps
		// the existing one-sample-after-bind rule, which restores its pose.
		e.sampled = c.sampled
	}
	if paused && c.paints != nil && !e.player.WantsRestamp() {
		writePaints(next, e.indices, c.paints)
	}
}

// restoreBases rewinds the paint a rebuilt player composes from when the
// display it was just bound to still carries the sample the previous player
// wrote. A re-projected display is built from the retained layout result, and
// the previous sample wrote through that same operation storage, so its
// operations show the sampled pose rather than the declared paint. Binding
// that pose as the base would compose the next sample on top of the last one:
// a paused chip translates a quarter cycle further every tick.
//
// The operation is compared with the pose the previous player left before its
// base is installed, so a display that really was rebuilt from the declared
// styles, or one a relayout moved, keeps the new paint. A player that needs the
// engine's re-projection pass writes no 2D matrix of its own, and its
// operations carry the engine's projection, so it is left alone.
func restoreBases(e *entry, c carry, next *Display) {
	if next == nil || e.paintMask == 0 || len(c.bases) == 0 || len(c.bases) != len(c.paints) ||
		len(c.bases) != len(e.indices) || e.player.WantsRestamp() {
		return
	}
	bases := e.player.BasePaints(nil)
	if len(bases) != len(c.bases) {
		return
	}
	changed := false
	for k, index := range e.indices {
		if index < 0 || index >= len(next.Ops) {
			return
		}
		if !samePose(Read(&next.Ops[index]), c.paints[k], e.paintMask) {
			continue
		}
		e.keptPose = true
		base := c.bases[k]
		base.Mask, base.NeedsRestamp = e.paintMask, false
		if bases[k] != base {
			bases[k], changed = base, true
		}
	}
	if changed {
		e.player.SetBasePaints(bases)
	}
}

// samePose reports whether two paints carry the same values in the fields mask
// names. It is how a rebind tells a display that already carries its own
// sample from one that was rebuilt from the declared styles.
func samePose(a, b Paint, mask Mask) bool {
	switch {
	case mask&MaskX != 0 && a.X != b.X:
		return false
	case mask&MaskY != 0 && a.Y != b.Y:
		return false
	case mask&MaskW != 0 && a.W != b.W:
		return false
	case mask&MaskH != 0 && a.H != b.H:
		return false
	case mask&MaskOpacity != 0 && a.Opacity != b.Opacity:
		return false
	case mask&MaskRGB != 0 && (a.R != b.R || a.G != b.G || a.B != b.B):
		return false
	case mask&MaskAlpha != 0 && a.Alpha != b.Alpha:
		return false
	case mask&MaskMatrix != 0 && (a.HasMatrix != b.HasMatrix || a.Matrix != b.Matrix):
		return false
	}

	return true
}

// writePaints copies a carried pose onto the operations the entry now targets.
// It runs after BindBox captured the target's base paint, so the player still
// composes from the declared paint while the op already shows the paused pose.
func writePaints(d *Display, indices []int, paints []Paint) {
	if d == nil {
		return
	}
	for k, index := range indices {
		if k >= len(paints) {
			break
		}
		if index < 0 || index >= len(d.Ops) {
			continue
		}
		Write(&d.Ops[index], paints[k])
	}
}

// runnable reports whether any listed animation starts on its own. A paused
// sample has a pose to show once, so it asks for that first frame and then
// parks; a triggered sample waits for an event and an unnamed sample has no
// keyframes to apply.
func runnable(items []layout.Animation) bool {
	for i := range items {
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
func opacityGap(d *Display, box layout.Box, bound Bound) string {
	if bound.Mask&MaskOpacity == 0 || len(bound.Indices) < 2 ||
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

// triggerAlignment maps each of the element's bound players to the animation
// records that produced its effects, so csstrigger.Dispatch can address one
// action by box index and animation index. BindBox merges the classes of
// several animations into one player, so a raw box.Animations slice would
// address the wrong effect; a per-animation bind recovers how many effects
// each animation contributed to each target set.
//
// It returns nil when the element declares no animation-trigger, which is the
// common case, so a page without triggers pays nothing. A returned gap string
// reports an element whose alignment failed; the animation still plays, it
// just cannot be triggered.
func triggerAlignment(d *Display, box layout.Box, boxIndex int, bounds []Bound,
	start time.Time, idx *Index) ([][]layout.Animation, string) {
	if !hasTrigger(box.Animations) || len(bounds) == 0 {
		return nil, ""
	}
	out := make([][]layout.Animation, len(bounds))
	offset := make([]int, len(bounds))
	for i := range bounds {
		count := bounds[i].Player.EffectCount()
		out[i] = make([]layout.Animation, count)
		for j := range out[i] {
			out[i][j].Trigger = "none"
		}
	}
	for i := range box.Animations {
		solo := box
		solo.Animations = box.Animations[i : i+1]
		soloBounds, _, err := BindBoxWithIndex(d, solo, boxIndex, start, idx)
		if err != nil {
			return nil, fmt.Sprintf("box %q: animation-trigger could not be aligned: %v", box.ID, err)
		}
		for j := range soloBounds {
			group := matchingBound(bounds, soloBounds[j].Indices)
			if group < 0 {
				continue
			}
			count := min(soloBounds[j].Player.EffectCount(), len(out[group])-offset[group])
			for k := 0; k < count; k++ {
				out[group][offset[group]+k].Trigger = box.Animations[i].Trigger
			}
			offset[group] += count
		}
	}
	for i := range out {
		if offset[i] != len(out[i]) {
			return nil, fmt.Sprintf("box %q: animation-trigger effect alignment is incomplete", box.ID)
		}
	}

	return out, ""
}

// matchingBound returns the bound group with the same target set, or -1.
func matchingBound(bounds []Bound, indices []int) int {
	for i := range bounds {
		if slices.Equal(bounds[i].Indices, indices) {
			return i
		}
	}

	return -1
}

// hasTrigger reports whether any listed animation declares a trigger name. A
// trigger without a keyframe rule still names a host event, but it binds no
// player of its own.
func hasTrigger(items []layout.Animation) bool {
	for i := range items {
		if items[i].Trigger != "" && items[i].Trigger != "none" {
			return true
		}
	}

	return false
}

// indexFor returns a cached or newly built *cssmotion.Index for d.
func (r *Runtime) indexFor(d *Display) *Index {
	if d == nil || len(d.Ops) == 0 {
		return EmptyIndex()
	}
	if r.index != nil && r.index != EmptyIndex() {
		if r.index.Display() == d {
			return r.index
		}
		if r.display == d || (len(d.Ops) == r.index.NumOps() && r.display != nil && sameKeyframes(d.Keyframes, r.display.Keyframes)) {
			r.index.Repoint(d)
			return r.index
		}
	}
	r.index = NewIndex(d)
	return r.index
}

func sameKeyframes(a, b []layout.KeyframesRule) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	return &a[0] == &b[0]
}

// BindAnimation builds an animation player from one box's Blinkless CSS metadata. It
// targets every captured operation with the full paint mask, which is the
// whole-subtree behaviour the catalog previews drive by hand. A page's
// automatic runtime uses BindBox instead, so a colour keyframe only writes the
// operations its property owns.
func BindAnimation(d *layout.Display, box layout.Box, indices []int, initial []Paint,
	start time.Time) (*Player, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, errors.New("cssanimation: box has no animation metadata")
	}
	ops := Resolve(d, indices)
	if len(ops) != len(indices) {
		return nil, errors.New("cssanimation: captured operations are missing from display")
	}
	if len(ops) == 0 || len(initial) != len(ops) {
		return nil, errors.New("cssanimation: target operation paints do not match")
	}
	targets := make([]Target, len(ops))
	for i, op := range ops {
		targets[i] = Target{
			Op:      op,
			From:    initial[i], //nolint:gosec // len(initial) is checked against len(ops) above
			Parent:  box.Transform.Parent,
			World:   box.Transform.World,
			OriginX: box.Transform.OriginX,
			OriginY: box.Transform.OriginY,

			HasElementTransform: box.Transform.HasMatrix,
			HasTransform3D:      box.Transform.HasTransform3D,
		}
	}
	size := transformSize{width: box.W, height: box.H, pointsPerPixel: displayScale(d), hasBox: true}
	base := perspectiveBaseFromTransform(size, box.Transform)
	effects := make([]Effect, 0, len(box.Animations))
	for _, item := range box.Animations {
		var frames []Keyframe
		if item.Name != "none" && item.Name != "" {
			frames = parseKeyframes(d.Keyframes, item.Name, size, nil, Color{}, false, base)
		}
		effects = append(effects, Effect{Frames: frames,
			Timing: timing(item), Trigger: item.Trigger})
	}
	player, err := NewComposite(targets, effects)
	if err != nil {
		return nil, err
	}
	player.Start(start)
	return player, nil
}

// Bound is one player bound to one target set: the operations it writes, the
// paint fields it may write, and the player itself.
type Bound struct {
	Player  *Player
	Indices []int
	Mask    Mask
}

// BindBox builds the players for one element's CSS animations, one player per
// distinct target set. boxIndex is the box's position in d.Boxes, or -1 when
// the element has no id the runtime can trust; an owner-scoped colour then
// falls back to the box's subtree and the binder reports the fallback.
//
// A property is split out of its animation by target set and field mask: a
// background colour writes only the element's own fill operations, a text
// colour writes only its own text operations, and a transform or opacity writes
// the whole subtree. Two classes with the same target set share one player, so
// two animations with disjoint fields still compose on one operation.
//
// Inline-SVG paint is not a display field: an animation whose only declarations
// are fill or stroke targets the SVG root's image operation with SVGPaintMask,
// so the player samples the paint without writing the operation. The page
// restyles the raster through layout.RestyleInlineSVG. A rule that also declares
// a paint class keeps the SVG sample on that player, because every class parses
// the same SVG declarations.
//
// A property this phase cannot write is left at the element's base paint and
// reported as a fidelity gap. The returned gaps are ordered and deduplicated
// per element and reason.
func BindBox(d *layout.Display, box layout.Box, boxIndex int, start time.Time) ([]Bound, []string, error) {
	return bindBox(d, box, boxIndex, start, nil)
}

// BindBoxWithIndex builds the players for one element's CSS animations,
// resolving every operation through an index the caller shares across boxes.
// It returns exactly what BindBox returns for the same display; sharing one
// index across a bind replaces a full operation scan per box per paint class
// with lookups. A nil index selects directly, as BindBox does.
func BindBoxWithIndex(d *layout.Display, box layout.Box, boxIndex int, start time.Time,
	idx *Index) ([]Bound, []string, error) {
	return bindBox(d, box, boxIndex, start, idx)
}

func bindBox(d *layout.Display, box layout.Box, boxIndex int, start time.Time,
	idx *Index) ([]Bound, []string, error) {
	if d == nil || len(box.Animations) == 0 {
		return nil, nil, errors.New("cssanimation: box has no animation metadata")
	}
	size := transformSize{width: box.W, height: box.H, pointsPerPixel: displayScale(d), hasBox: true}
	base := perspectiveBaseFromTransform(size, box.Transform)
	var report gapList

	current, hasCurrent := Color{}, false
	for i := range box.Animations {
		if usesCurrentColor(ruleFor(d.Keyframes, box.Animations[i].Name)) {
			current, hasCurrent = elementColor(d, box, boxIndex, idx)
			break
		}
	}

	var groups []targetGroup
	for i := range box.Animations {
		item := box.Animations[i]
		source := ruleFor(d.Keyframes, item.Name)
		if len(source) == 0 {
			continue
		}
		reportDeferred(&report, box, source)
		classes := effectClasses(source)
		if len(classes) == 0 && declaresSVGPaint(source) {
			classes = []paintClass{svgPaintClass()}
		}
		for _, class := range classes {
			frames := keyframesForBase(source, size, class, current, hasCurrent, base)
			if len(frames) == 0 {
				continue
			}
			if err := frameError(frames); err != nil {
				return nil, report.list, fmt.Errorf("keyframe %s: %w", class.name(), err)
			}
			reportSVGPaintDeferred(&report, box, frames)
			var indices []int
			if class.mask == SVGPaintMask {
				if box.ID == "" {
					report.add(fmt.Sprintf(
						"box %q: inline-svg paint needs an element id", box.ID))
					continue
				}
				if box.Tag != "svg" {
					report.add(fmt.Sprintf(
						"box %q: <%s> does not paint an inline svg", box.ID, box.Tag))
					continue
				}
				indices = svgImageIndices(d, box, boxIndex, idx)
				if len(indices) == 0 {
					report.add(fmt.Sprintf(
						"box %q: inline-svg paint found no svg raster", box.ID))
					continue
				}
			} else {
				if class.kind != KindSubtree && boxIndex < 0 {
					report.add(fmt.Sprintf(
						"box %q has no display index; owner-scoped colours fall back to its subtree", box.ID))
				}
				if idx != nil {
					indices = idx.Select(box, boxIndex, class.kind)
				} else {
					indices = Select(d, box, boxIndex, class.kind)
				}
				if len(indices) == 0 {
					continue
				}
			}
			reportColorAlpha(&report, box, frames)
			group := groupFor(&groups, indices)
			group.mask |= class.mask
			group.effects = append(group.effects, Effect{
				Frames: frames, Timing: timing(item), Trigger: item.Trigger,
			})
		}
	}

	bounds := make([]Bound, 0, len(groups))
	for i := range groups {
		targets, err := groups[i].targets(d, box)
		if err != nil {
			return nil, report.list, err
		}
		player, err := NewComposite(targets, groups[i].effects)
		if err != nil {
			return nil, report.list, err
		}
		player.Start(start)
		bounds = append(bounds, Bound{Player: player, Indices: groups[i].indices, Mask: groups[i].mask})
	}
	return bounds, report.list, nil
}

// targetGroup collects the effects of one element that write the same target
// set, so they sample through one player and compose in declaration order.
type targetGroup struct {
	indices []int
	mask    Mask
	effects []Effect
}

// groupFor finds the group already targeting this exact set, or appends a new
// one.
func groupFor(groups *[]targetGroup, indices []int) *targetGroup {
	for i := range *groups {
		if slices.Equal((*groups)[i].indices, indices) {
			return &(*groups)[i]
		}
	}
	*groups = append(*groups, targetGroup{indices: indices})
	return &(*groups)[len(*groups)-1]
}

// targets captures one operation per index and stamps this group's field mask
// on its starting paint. The mask is what the player writes with: a transform
// group never touches colour, and a colour group never touches the matrix.
func (g targetGroup) targets(d *layout.Display, box layout.Box) ([]Target, error) {
	targets := make([]Target, 0, len(g.indices))
	for _, index := range g.indices {
		if index < 0 || index >= len(d.Ops) {
			return nil, fmt.Errorf("cssanimation: operation %d is missing from display", index)
		}
		from := Read(&d.Ops[index])
		from.Mask = g.mask

		targets = append(targets, Target{
			Op:      &d.Ops[index],
			From:    from,
			Parent:  box.Transform.Parent,
			World:   box.Transform.World,
			OriginX: box.Transform.OriginX,
			OriginY: box.Transform.OriginY,

			HasElementTransform: box.Transform.HasMatrix,
			HasTransform3D:      box.Transform.HasTransform3D,
		})
	}
	return targets, nil
}

// elementColor resolves currentColor from the element's own text operation. The
// engine does not export the computed colour on the box, so a box whose text is
// painted by something else reports false and a currentColor keyframe becomes a
// bind error.
func elementColor(d *layout.Display, box layout.Box, boxIndex int, idx *Index) (Color, bool) {
	var indices []int
	if idx != nil {
		indices = idx.Select(box, boxIndex, KindInk)
	} else {
		indices = Select(d, box, boxIndex, KindInk)
	}
	for _, index := range indices {
		op := &d.Ops[index]
		return Color{R: op.R, G: op.G, B: op.B, A: maxChannel}, true
	}
	return Color{}, false
}

// usesCurrentColor reports whether a rule animates any colour to or from
// currentColor. Only then does the binder need the element's own ink colour.
func usesCurrentColor(source []layout.Keyframe) bool {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			if _, ok := propertyClass(declaration.Property); !ok {
				continue
			}
			if strings.EqualFold(strings.TrimSpace(declaration.Value), "currentcolor") {
				return true
			}
		}
	}
	return false
}

// reportDeferred counts the keyframe properties that belong to a later phase.
// The declarations stay at the element's base paint, so the animation runs but
// that property does not move. Inline-SVG paint is not one of them any more:
// the binder routes it to the restyle path, and a paint server it cannot use is
// reported by reportSVGPaintDeferred instead.
func reportDeferred(report *gapList, box layout.Box, source []layout.Keyframe) {
	for i := range source {
		for _, declaration := range source[i].Declarations {
			if isSVGPaintProperty(declaration.Property) {
				continue
			}
			phase, ok := deferredPhase(declaration.Property)
			if !ok {
				continue
			}
			report.add(fmt.Sprintf("box %q: %s is deferred to phase %02d",
				box.ID, declaration.Property, phase))
		}
	}
}

// reportColorAlpha counts a keyframed colour whose alpha the paint phase cannot
// write. The sRGB channels animate; the composite alpha stays at the element's
// base paint until an alpha field exists.
func reportColorAlpha(report *gapList, box layout.Box, frames []Keyframe) {
	for i := range frames {
		if !frames[i].ColorAlphaSet {
			continue
		}
		report.add(fmt.Sprintf("box %q: keyframed colour alpha is left at base paint", box.ID))
		return
	}
}

// gapList collects fidelity gaps once each, in the order they appear.
type gapList struct {
	list []string
	seen map[string]struct{}
}

func (g *gapList) add(reason string) {
	if g.seen == nil {
		g.seen = make(map[string]struct{})
	}
	if _, ok := g.seen[reason]; ok {
		return
	}
	g.seen[reason] = struct{}{}
	g.list = append(g.list, reason)
}

// displayScale returns the display's point-to-pixel factor, or 1 for a
// hand-built display that left it zero. A zero factor would collapse every
// percentage transform length to zero.
func displayScale(d *layout.Display) float64 {
	if d != nil && d.PointsPerPixel > 0 {
		return d.PointsPerPixel
	}
	return 1
}

// paintClass is one property family an effect writes: the operations the
// property targets and the paint fields it fills. Two properties with the same
// target set but a different field mask stay separate classes, and two colour
// properties with different target kinds (a background fill against text ink)
// stay separate even though both write RGB.
type paintClass struct {
	kind Kind
	mask Mask
}

// propertyClass maps a keyframed property to the operations and paint fields
// it writes. ok is false for every property this phase leaves alone, including
// the ones deferredPhase names.
func propertyClass(property string) (paintClass, bool) {
	switch property {
	case "transform", "perspective", "perspective-origin":
		return paintClass{kind: KindSubtree, mask: MaskMatrix}, true
	case "opacity":
		return paintClass{kind: KindSubtree, mask: MaskOpacity}, true
	case "background-color", "accent-color":
		return paintClass{kind: KindFill, mask: MaskRGB}, true
	case "color", "text-decoration-color", "text-emphasis-color":
		return paintClass{kind: KindInk, mask: MaskRGB}, true
	case "border-color", "border-top-color", "border-right-color",
		"border-bottom-color", "border-left-color", "border-block-color",
		"border-block-start-color", "border-block-end-color", "border-inline-color",
		"border-inline-start-color", "border-inline-end-color", "outline-color",
		"column-rule-color", "row-rule-color", "rule-color":
		return paintClass{kind: KindStroke, mask: MaskRGB}, true
	}
	return paintClass{}, false
}

// deferredPhase reports the phase that owns a keyframe property this phase
// leaves at base paint. The binder counts one gap per element and property, so
// a deferred declaration is reported rather than dropped silently.
// Deferred property phases. The number is the phase that owns the property, so
// a gap report can point a reader at the right plan.
const (
	phaseTransform = 2
	phaseLayout    = 4
	phaseFilter    = 8
)

// deferredPhase names the phase that owns a property this binder leaves at the
// element's base paint. Inline-SVG paint is not here: the parser samples it in
// svg_values.go and BindBox routes the sample to layout.RestyleInlineSVG.
func deferredPhase(property string) (int, bool) {
	switch property {
	case "filter", "backdrop-filter", "clip-path", "mask-position", "mask-size",
		"mix-blend-mode", "isolation":
		return phaseFilter, true
	case "translate", "rotate", "scale":
		return phaseTransform, true
	}
	if deferredLayoutProperty(property) {
		return phaseLayout, true
	}
	return 0, false
}

// keyframes parses one rule's declarations for every paint class. The binder
// parses one class at a time through keyframesForBase; this form keeps the
// per-declaration view the transform and perspective tests use.
func keyframes(rules []layout.KeyframesRule, name string, pointsPerPixel, width, height float64) []Keyframe {
	size := transformSize{width: width, height: height, pointsPerPixel: pointsPerPixel, hasBox: true}
	return parseKeyframes(rules, name, size, nil, Color{}, false, PerspectiveBase{})
}

// keyframesFor parses one rule's declarations for one paint class. A
// declaration of another class is skipped, so the frames carry only the fields
// this class writes: a background class never sees the transform, and a text
// class never sees the background colour.
func keyframesFor(source []layout.Keyframe, size transformSize,
	class paintClass, current Color, hasCurrent bool) []Keyframe {
	return parseFrames(source, size, &class, current, hasCurrent, PerspectiveBase{})
}

// keyframesForBase is keyframesFor with the element's static perspective and
// vanishing point, so an implicit endpoint holds the element's static camera.
func keyframesForBase(source []layout.Keyframe, size transformSize,
	class paintClass, current Color, hasCurrent bool, base PerspectiveBase) []Keyframe {
	return parseFrames(source, size, &class, current, hasCurrent, base)
}

// parseKeyframes parses every declaration of one rule. A nil class keeps every
// property, which is the unfiltered view the legacy bind uses.
func parseKeyframes(rules []layout.KeyframesRule, name string, size transformSize,
	class *paintClass, current Color, hasCurrent bool,
	base PerspectiveBase) []Keyframe {
	return parseFrames(ruleFor(rules, name), size, class, current, hasCurrent, base)
}

func parseFrames(source []layout.Keyframe, size transformSize,
	class *paintClass, current Color, hasCurrent bool,
	base PerspectiveBase) []Keyframe {
	frames := make([]Keyframe, len(source))
	for i, item := range source {
		frame := Keyframe{
			Offset:          item.Offset,
			Perspective:     framePerspective(size, base),
			PerspectiveBase: base,
		}
		for _, declaration := range item.Declarations {
			// A keyframe timing function is not a paint property, so it is
			// read before the paint-class filter: every paint class of the
			// rule eases its segment with the same declared function.
			if keyframeTimingFunction(declaration) {
				frame.TimingFunction = strings.TrimSpace(declaration.Value)
				continue
			}
			// Inline-SVG paint targets no display operation, so it is read
			// before the paint-class filter: every class of the rule carries
			// the same sampled paint for the element's SVG box.
			if isSVGPaintProperty(declaration.Property) {
				parseSVGPaintDeclaration(&frame, declaration, size, current, hasCurrent)
				continue
			}
			pc, ok := propertyClass(declaration.Property)
			if !ok || (class != nil && pc != *class) {
				continue
			}
			parseDeclaration(&frame, declaration, size, current, hasCurrent)
		}
		frames[i] = frame
	}
	sort.SliceStable(frames, func(i, j int) bool { return frames[i].Offset < frames[j].Offset })
	return frames
}

// keyframeTimingFunction reports a declaration that eases the interval starting
// at this keyframe. It belongs to no paint class, so the parser handles it for
// every class. The vendor-prefixed form is the same longhand alias the engine
// accepts outside a keyframe block.
func keyframeTimingFunction(declaration layout.KeyframeDeclaration) bool {
	property := strings.TrimSpace(declaration.Property)
	return strings.EqualFold(property, "animation-timing-function") ||
		strings.EqualFold(property, "-webkit-animation-timing-function")
}

// parseDeclaration writes one declaration into a frame. A value this phase
// cannot parse is kept as a bind error; a colour space a later phase owns is
// left at the element's base paint and counted by the binder.
func parseDeclaration(frame *Keyframe, declaration layout.KeyframeDeclaration,
	size transformSize, current Color, hasCurrent bool) {
	switch declaration.Property {
	case "transform":
		value, err := parseTransform(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Transform = value
		frame.X, frame.Y = value.Decomposed.Tx, value.Decomposed.Ty
		frame.Move = true
	case "perspective":
		distance, err := parsePerspective(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Perspective.Distance = distance
		frame.Perspective.HasDistance = true
	case "perspective-origin":
		x, y, err := parsePerspectiveOrigin(declaration.Value, size)
		if err != nil {
			frame.Err = err
			return
		}
		frame.Perspective.OriginX, frame.Perspective.OriginY = x, y
		frame.Perspective.HasOrigin = true
	case "opacity":
		value, ok := number(declaration.Value)
		if !ok {
			frame.Err = fmt.Errorf("%w: opacity %q", ErrNumber, declaration.Value)
			return
		}
		frame.Opacity, frame.OpacitySet = value, true
	default:
		value, err := resolveColor(declaration.Value, current, hasCurrent)
		if err != nil {
			if !errors.Is(err, ErrColorSpace) {
				frame.Err = err
			}
			return
		}
		frame.R, frame.G, frame.B, frame.ColorSet = value.R, value.G, value.B, true
		if value.A < maxChannel {
			frame.ColorAlpha, frame.ColorAlphaSet = value.A, true
		}
	}
}

// name returns the property family a class stands for, for error text.
func (c paintClass) name() string {
	if c.mask == MaskOpacity {
		return "opacity"
	}
	switch c.kind {
	case KindFill:
		return "fill colour"
	case KindInk:
		return "text colour"
	case KindStroke:
		return "stroke colour"
	}
	return "transform"
}

// frameError returns the first declaration error in a parsed frame list. The
// binder reports it even when no operation would be targeted, so a bad
// declaration is never dropped because the element paints the wrong kind.
func frameError(frames []Keyframe) error {
	for i := range frames {
		if frames[i].Err != nil {
			return fmt.Errorf("frame %d: %w", i, frames[i].Err)
		}
	}
	return nil
}

// resolveColor parses one colour declaration and resolves currentColor against
// the element's own ink colour.
func resolveColor(value string, current Color, hasCurrent bool) (Color, error) {
	parsed, err := parseColor(value)
	if err != nil {
		return Color{}, err
	}
	if !parsed.Current {
		return parsed, nil
	}
	if !hasCurrent {
		return Color{}, ErrCurrentColor
	}
	return current, nil
}

// ruleFor returns the keyframe declarations of the last rule with this name,
// which is the rule CSS selects when a name repeats.
func ruleFor(rules []layout.KeyframesRule, name string) []layout.Keyframe {
	for i := len(rules) - 1; i >= 0; i-- {
		if rules[i].Name == name {
			return rules[i].Keyframes
		}
	}
	return nil
}

// effectClasses returns the distinct paint classes the declarations of one
// effect use, in declaration order.
func effectClasses(source []layout.Keyframe) []paintClass {
	var classes []paintClass
	for i := range source {
		for _, declaration := range source[i].Declarations {
			class, ok := propertyClass(declaration.Property)
			if !ok || containsClass(classes, class) {
				continue
			}
			classes = append(classes, class)
		}
	}
	return classes
}

func containsClass(classes []paintClass, class paintClass) bool {
	for _, candidate := range classes {
		if candidate == class {
			return true
		}
	}
	return false
}

func timing(a layout.Animation) Timing {
	return Timing{Duration: ms(a.DurationMS), Delay: ms(a.DelayStartMS),
		EndDelay: ms(a.DelayEndMS), Iterations: a.IterationCount, Infinite: a.IterationInfinite,
		Direction: direction(a.Direction), Fill: fill(a.FillMode), Play: parsePlayState(a.PlayState),
		TimingFunction: a.TimingFunction, DurationAuto: a.DurationAuto,
		Composition: composition(a.Composition), Timeline: a.Timeline,
		RangeStart: a.RangeStart, RangeEnd: a.RangeEnd, RangeCenter: a.RangeCenter}
}

func composition(value string) Composition {
	switch value {
	case "add":
		return Add
	case "accumulate":
		return Accumulate
	default:
		return Replace
	}
}

func ms(value float64) time.Duration { return time.Duration(value * float64(time.Millisecond)) }

func direction(value string) Direction {
	switch value {
	case "reverse":
		return Reverse
	case "alternate":
		return Alternate
	case "alternate-reverse":
		return AlternateReverse
	default:
		return Normal
	}
}

func fill(value string) FillMode {
	switch value {
	case "forwards":
		return FillForwards
	case "backwards":
		return FillBackwards
	case "both":
		return FillBoth
	default:
		return FillNone
	}
}

func parsePlayState(value string) PlayState {
	if value == "paused" {
		return Paused
	}
	return Running
}
