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
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
	"time"
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

	// index is the cached operation lookup for display.
	index *Index

	entries []entry
	stats   Stats
	// promoted holds the element ids the compositor kept as retained layers
	// on the last frame. Step leaves their page damage to the layer cache.
	promoted map[string]struct{}
	// promotedBoxes holds the box indices the compositor kept as retained
	// layers on the last frame, so anonymous promoted layers suppress
	// background damage by boxIndex as well as elementID.
	promotedBoxes map[int]struct{}

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

	// Scratch maps reused by CaptureTransitions across frames so steady-state
	// commits avoid rebuilding per-element lookup maps on every hover.
	capScope    map[string]struct{}
	capIDs      map[string]struct{}
	capOrdinals map[string]int
	capOldBoxes map[entryKey]boxRef
	capRecords  map[string]layout.Transition

	// now is the clock a newly bound effect starts at. Tests replace it.
	now func() time.Time
}

// New returns an idle runtime with no players bound.
func New() *Runtime {
	return &Runtime{now: time.Now}
}

// viewportSlack keeps a card just above or below the scrollport warm, so a
// short scroll does not sample a card for the first time as it appears.
const viewportSlack = 96.0

// SetPromotedLayers tells the runtime which elements the compositor kept as
// retained layers on the last frame. A promoted element paints into its layer
// texture, so Step stops adding page background damage for it. A demotion
// repairs the pose through the renderer's layer topology damage.
func (r *Runtime) SetPromotedLayers(ids []string) {
	if r == nil {
		return
	}
	if len(ids) == 0 {
		r.promoted = nil

		return
	}

	if r.promoted == nil {
		r.promoted = make(map[string]struct{}, len(ids))
	} else {
		clear(r.promoted)
	}
	for _, id := range ids {
		if id != "" {
			r.promoted[id] = struct{}{}
		}
	}
}

// SetPromotedLayerBoxes tells the runtime which box indices the compositor
// kept as retained layers on the last frame, so anonymous promoted layers can
// suppress background damage by boxIndex as well as elementID.
func (r *Runtime) SetPromotedLayerBoxes(boxes []int) {
	if r == nil {
		return
	}
	if len(boxes) == 0 {
		r.promotedBoxes = nil

		return
	}

	if r.promotedBoxes == nil {
		r.promotedBoxes = make(map[int]struct{}, len(boxes))
	} else {
		clear(r.promotedBoxes)
	}
	for _, box := range boxes {
		if box >= 0 {
			r.promotedBoxes[box] = struct{}{}
		}
	}
}

// promotedBox reports whether boxIndex is currently promoted to a retained
// layer.
func (r *Runtime) promotedBox(boxIndex int) bool {
	if r == nil || boxIndex < 0 || len(r.promotedBoxes) == 0 {
		return false
	}

	_, ok := r.promotedBoxes[boxIndex]

	return ok
}

// promotedEntry reports that the entry's changed paint lands in a retained
// layer, so it needs no background repaint. A layout entry on the same
// element still moves the page and keeps its damage.
func (r *Runtime) promotedEntry(e *entry) bool {
	if r == nil || e.layout || e.svg {
		return false
	}
	if e.key.id != "" && len(r.promoted) > 0 {
		if _, ok := r.promoted[e.key.id]; ok {
			return true
		}
	}

	return r.promotedBox(e.boxIndex)
}

// promotedTransition reports whether a live transition's element is currently
// promoted to a retained layer (by element id or box index), so its sample
// needs no background repaint.
func (r *Runtime) promotedTransition(d *Display, e *transitionEntry) bool {
	if r == nil {
		return false
	}
	if e.key.element.id != "" && len(r.promoted) > 0 {
		if _, ok := r.promoted[e.key.element.id]; ok {
			return true
		}
	}
	if len(r.promotedBoxes) > 0 && d != nil {
		if box := transitionBoxIndex(d, e.key.element); box >= 0 {
			return r.promotedBox(box)
		}
	}

	return false
}

// visible reports whether the entry's border box is in or near the page's
// scroll viewport. A runtime with no view keeps every entry visible.
func (r *Runtime) visible(e *entry) bool {
	if r == nil || !r.hasView {
		return true
	}

	v := r.view

	return e.boxX+e.boxW >= v.ScrollX-viewportSlack &&
		e.boxX <= v.ScrollX+v.Width+viewportSlack &&
		e.boxY+e.boxH >= v.ScrollY-viewportSlack &&
		e.boxY <= v.ScrollY+v.Height+viewportSlack
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

// CheckManualOverlap reports whether a manual tick callback mutated any
// display operation currently owned by an active automatic CSS animation or
// transition. When an overlap is found, it increments Stats.ManualOverlap and
// records Stats.ManualOverlapID so callers and DevTools can diagnose
// dual-driven operations without logging on the hot path.
func (r *Runtime) CheckManualOverlap(before, after []layout.DisplayOp) (string, bool) {
	if r == nil || !r.Active() || len(before) != len(after) {
		return "", false
	}
	for i := range r.entries {
		e := &r.entries[i]
		if !e.timed || e.layout {
			continue
		}
		if (e.sampled || e.idle) && (e.player == nil || !e.player.Active()) {
			continue
		}
		if opsMutated(before, after, e.indices) {
			r.stats.ManualOverlap++
			r.stats.ManualOverlapID = e.key.id
			return e.key.id, true
		}
	}
	for i := range r.transitions {
		t := &r.transitions[i]
		if t.sampled && !t.active {
			continue
		}
		if opsMutated(before, after, t.indices) {
			r.stats.ManualOverlap++
			r.stats.ManualOverlapID = t.key.element.id
			return t.key.element.id, true
		}
	}
	return "", false
}

func opsMutated(before, after []layout.DisplayOp, indices []int) bool {
	for _, idx := range indices {
		if idx < 0 || idx >= len(before) {
			continue
		}
		a, b := &before[idx], &after[idx]
		if a.X != b.X || a.Y != b.Y || a.W != b.W || a.H != b.H ||
			a.R != b.R || a.G != b.G || a.B != b.B || a.Alpha != b.Alpha ||
			a.Opacity() != b.Opacity() || a.XformValue() != b.XformValue() {
			return true
		}
	}
	return false
}

// NextFrame returns the next time the runtime needs a sample. An animation and
// a running transition ask for the next frame; a transition still inside its
// delay asks for the delay end instead, so a delayed page parks rather than
// repainting the same starting paint every frame. ok is false when nothing is
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
	Players map[int]*Player
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
		players := make(map[int]*Player, len(byBox))
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

// Stats is a snapshot of the runtime counters. The page copies it into the
// probe block; motion logs nothing on the hot path.
type Stats struct {
	// Bound is the number of players the last bind installed.
	Bound int
	// Active is the number of players that reported work at the last Step.
	Active int
	// Sampled is the number of players the last Step sampled.
	Sampled int
	// NeedsRestamp is the number of sampled players whose transform or
	// perspective could not be written as a 2D matrix at the last Step,
	// because the element or the keyframe is 3D, or because it sampled a
	// perspective. Step collects one Request per element for them and the
	// page applies them through layout.RestampTransforms.
	NeedsRestamp int
	// NeedsRetainedState reports whether the last bind found motion that needs
	// the re-projection pass. Without retained layout state the engine pass
	// cannot run, so the page turns it on and lays the page out once more
	// before the first restamp.
	NeedsRetainedState bool
	// Restamped is the number of elements the page re-projected at the last
	// tick.
	Restamped int
	// RestampError is the last re-projection failure. The page keeps drawing
	// the display it already had, so a bad override leaves the last good frame
	// on screen and the next sample tries again.
	RestampError string
	// Rebinds counts the binds that replaced at least one bound player.
	Rebinds int
	// LastError is the last bind failure. It is empty when the last bind
	// succeeded; a failed box does not stop the other boxes from binding.
	LastError string
	// Gaps lists the fidelity gaps the last bind accepted, one entry per
	// element and reason. A gap is a property or approximation the runtime
	// knows it cannot reproduce exactly yet: the declaration stays at the
	// element's base paint, or the approximation is written and reported, such
	// as per-operation opacity on overlapping content until phase 8. The list
	// is rebuilt at every bind and empty after Bind(nil).
	Gaps []string
	// UnresolvedTimelines counts the dashed-ident animation-timeline
	// references the last bind could not resolve: a name no ancestor
	// declares, a nested element scroller this engine does not implement, or
	// a view timeline whose subject is not in the display. Those animations
	// stay inactive; they never fall back to wall-clock time.
	UnresolvedTimelines int
	// InactiveTimelines counts the bound players the last Step left inactive
	// because a resolved named timeline had no range at that page geometry,
	// such as a page that does not scroll. It is reset by every Step.
	InactiveTimelines int
	// LayoutBound is the number of elements whose layout-tier keyframes the
	// last bind installed. Each one reaches the engine's override relayout
	// path through the per-frame budget.
	LayoutBound int
	// LayoutRelayouts is the number of elements the page relaid out through
	// the engine override pass at the last tick.
	LayoutRelayouts int
	// LayoutQueued is the number of elements still waiting for a relayout
	// after the last tick's budget. A nonzero value means the budget deferred
	// work; the elements sample when their turn runs, not when they queued.
	LayoutQueued int
	// LayoutBudgetHit reports that the last tick left at least one element
	// queued because the budget was spent. A page that keeps hitting it wants
	// a larger budget or a lower sample rate for layout animations.
	LayoutBudgetHit bool
	// LayoutError is the last override relayout failure. The page keeps
	// drawing the display it already had, so a bad value leaves the last good
	// frame on screen and the next queued sample tries again.
	LayoutError string
	// SVGPaintRequests is the number of inline-SVG paint values the last Step
	// collected, one per element whose paint changed. Each one is a candidate
	// for layout.RestyleInlineSVG.
	SVGPaintRequests int
	// SVGPaintRestyles is the number of elements the page restyled at the last
	// tick. It is one at most unless the budget was raised.
	SVGPaintRestyles int
	// SVGPaintQueued is the number of collected restyles the last take left
	// waiting for the per-frame budget. The next Step collects them again, so
	// they are never lost.
	SVGPaintQueued int
	// SVGPaintBudgetHit reports that the last take left at least one element
	// queued because the budget was spent. A page that keeps hitting it wants a
	// larger budget or fewer animated icons.
	SVGPaintBudgetHit bool
	// SVGPaintDeferred counts the samples whose paint property could not be
	// restyled, such as a url(#gradient) paint server. The element keeps its
	// base paint and SVGPaintDeferredReason names the declaration.
	SVGPaintDeferred int
	// SVGPaintDeferredReason is the first deferral reason of the last Step.
	SVGPaintDeferredReason string
	// SVGPaintError is the last restyle failure. The page keeps drawing the
	// display it already had, and the next tick collects the request again.
	SVGPaintError string
	// TransitionsBound is the number of transition renderers the last bind
	// installed.
	TransitionsBound int
	// TransitionsStarted is the number of new renderers the last bind started
	// from a captured style change.
	TransitionsStarted int
	// TransitionsCancelled is the number of live renderers the last bind
	// dropped: a removed element, a removed declaration, or a replaced run.
	TransitionsCancelled int
	// TransitionsSkipped is the number of captures or live entries the last
	// bind could not use because the element's operation count or kinds
	// changed, or because an animation owns the field. Those properties snap
	// to the committed value.
	TransitionsSkipped int
	// TransitionsBlocked is the number of property changes the last capture
	// saw but could not transition because the display does not expose the
	// value or the engine path cannot express it yet. A blocked change snaps
	// to the committed value.
	TransitionsBlocked int
	// TransitionSkipReason names the first property change the last capture
	// left alone with a reason: an outgoing presence with no destination
	// operations, a display cross-fade that needs the phase 04 layer, an
	// integer that needs a relayout property the engine does not expose, or a
	// layout length whose computed value is not in the display. It is empty
	// when nothing was blocked.
	TransitionSkipReason string
	// TransitionsNoID counts the transition declarations the last capture
	// skipped because the element has no id to match across displays.
	TransitionsNoID int
	// TransitionsNoOld counts the declarations the last capture skipped
	// because the element was not in the previous display, which is an
	// appearing element rather than a style change.
	TransitionsNoOld int
	// TransitionsActive is the number of renderers that reported work at the
	// last step.
	TransitionsActive int
	// TransitionsSampled is the number of renderers the last step sampled.
	TransitionsSampled int
	// TransitionError is the last bind failure for a captured transition. The
	// other transitions still bind; a failed one snaps to the committed paint.
	TransitionError string
	// ManualOverlap counts how many times a Page.SetTick callback mutated a
	// display operation currently owned by an active automatic CSS animation
	// or transition. Turn off SetAnimations(false) or SetTransitions(false)
	// before driving those elements manually.
	ManualOverlap int
	// ManualOverlapID is the element ID of the most recent manual/CSS motion
	// overlap detected during a tick.
	ManualOverlapID string
}

// fieldMask names the display values one target may write. BindBox narrows it
// from the keyframe properties a player carries, so a transform-only animation
// stops rewriting paint and a colour animation stops writing the matrix.
type fieldMask uint16

// One bit per field a keyframe can change.
const (
	fieldX fieldMask = 1 << iota
	fieldY
	fieldMatrix
	fieldOpacity
	fieldRGB
)

// maskAll is every field a target can carry. A zero mask writes nothing.
const maskAll = fieldX | fieldY | fieldMatrix | fieldOpacity | fieldRGB

// fieldMaskOf narrows the driver's mask to the paint fields a player owns. A
// full snapshot maps to maskAll; W, H, and Alpha are not driver write fields,
// so they never appear here.
func fieldMaskOf(mask Mask) fieldMask {
	if mask == 0 || mask == MaskAll {
		return maskAll
	}
	var out fieldMask
	if mask&MaskX != 0 {
		out |= fieldX
	}
	if mask&MaskY != 0 {
		out |= fieldY
	}
	if mask&MaskMatrix != 0 {
		out |= fieldMatrix
	}
	if mask&MaskOpacity != 0 {
		out |= fieldOpacity
	}
	if mask&MaskRGB != 0 {
		out |= fieldRGB
	}
	return out
}

// geometry reports whether a mask can move or resize a target, so a sample
// only re-measures the painted bounds when the geometry may have changed.
func (m fieldMask) geometry() bool {
	return m&(fieldX|fieldY|fieldMatrix) != 0
}
