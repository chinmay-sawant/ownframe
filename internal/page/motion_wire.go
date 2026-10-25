package page

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// MotionStats is one snapshot of the automatic CSS motion counters and the
// last bind error. Probe().Motion returns it; the page copies the runtime
// block instead of logging on the hot path.
type MotionStats = motion.Stats

// bindMotion resolves the CSS animations of the display the page just
// published. Every publish path calls it after the display is installed and
// the generation bumped, so the runtime never samples an operation list the
// page has replaced. previous is the list the page drew before it, which the
// runtime reads for the pose a paused animation must keep across the rebind. A
// fallback page has no patchable new list, so it passes nil as display.
//
// A bind that finds a 3D transform keyframe, a 3D element, a perspective
// keyframe, or a layout-tier keyframe turns retained layout state on, because
// the re-projection and override relayout passes cannot run without it, and
// owes one redraw of the display the pass will patch. The frame the flag lands
// on does not move; the next tick applies the pass. A later publish of a
// retained display pays that debt, so a page that drew itself again before the
// tick does not wait an extra frame.
func (p *Page) bindMotion(previous, display *Display) {
	if p.motion == nil {
		return
	}

	if p.animationsOff {
		_ = p.motion.Bind(nil)

		return
	}

	// Every publish laid the display out with the retained state the page had
	// at draw time, so the flag before the bind says whether this one carries
	// the box tree the re-projection pass patches.
	retained := p.retainState

	_ = p.motion.BindAfter(previous, display)
	if display == nil {
		return
	}

	if p.motion.NeedsRetainedState() && !p.retainState {
		p.retainState = true
		p.restampRedraw = true

		return
	}

	if retained {
		p.restampRedraw = false
	}

	// A publish rebuilt the operations from the static styles, so a layout
	// animation that had already sampled (paused or finished) has to run again
	// to put its pose on the new display. The relayout adoption skips this,
	// because the display it bound is the one the relayout just built.
	if !p.layoutAdopting {
		p.motion.RequeueLayout()
	}
}

// captureTransitions reads the style change between the display the page drew
// and the one it is about to install. It must run before the commit swaps the
// old list away, because the old operations hold the paint currently on
// screen. A page with the transition driver off captures nothing.
func (p *Page) captureTransitions(previous, next *Display) []motion.TransitionCapture {
	if p.motion == nil || p.transitionsOff {
		return nil
	}
	return p.motion.CaptureTransitions(previous, next)
}

// bindTransitions installs the captures against the display the page just
// committed. The commit time is the transition start. When the page has no
// patchable display, or the switch is off, any live renderer is released and
// the committed paint stays as it is.
func (p *Page) bindTransitions(captures []motion.TransitionCapture) {
	if p.motion == nil {
		return
	}
	if p.transitionsOff {
		p.motion.ReleaseTransitions()

		return
	}
	p.motion.BindTransitions(p.draw.Display, captures, time.Now())
}

// stepMotion samples the layout pass first, then the transition pass, then the
// paint and matrix passes against the display the relayout produced, marks
// their damage, and asks for the next frame the runtime needs. Transitions step
// before animations, so an animation on the same field writes second and wins.
// The pass runs before the user tick callback, so the callback's snapshot
// already carries the sampled paint and its diff never covers a motion change.
func (p *Page) stepMotion(ctx context.Context) {
	if p.motion == nil || (p.animationsOff && p.transitionsOff) {
		return
	}

	p.setMotionView()
	now := time.Now()
	var damage motion.Damage
	if p.animationsOff {
		// A geometry transition reaches the engine through the same relayout
		// pass as a layout keyframe, so the pass runs even when the animation
		// driver is off.
		p.stepLayoutMotion(ctx, now)
		damage = p.motion.StepTransitions(p.draw.Display, now)
	} else {
		p.stepLayoutMotion(ctx, now)
		damage = p.motion.Step(p.draw.Display, now)
		p.stepSVGPaintMotion(ctx)
	}
	if damage.Overflow {
		p.markFull()
	} else {
		for i := 0; i < damage.N; i++ {
			p.markRect(damage.Rects[i])
		}
	}

	if at, ok := p.motion.NextFrame(now); ok {
		p.WantFrameAt(at)
	}
}

// stepLayoutMotion samples at most one queued layout animation and applies the
// engine patch before the paint pass runs, so a transform on the same or a
// neighbouring element composes against the boxes the relayout produced.
//
// The patch needs the retained layout state a display only carries when it was
// laid out with RetainState. A page that animated a layout property without it
// enables the state here and owes one redraw, the same debt the re-projection
// pass pays; the sample goes back to the head of the queue and runs on the
// frame that can apply it.
func (p *Page) stepLayoutMotion(ctx context.Context, now time.Time) {
	requests := p.motion.StepLayout(now)
	if len(requests) == 0 {
		return
	}

	if p.draw.Display == nil || p.cache == nil {
		p.motion.ReturnLayout(requests)

		return
	}

	if !p.retainState || p.restampRedraw {
		p.retainState = true
		p.restampRedraw = true
		p.motion.ReturnLayout(requests)

		return
	}

	p.applyLayout(ctx, requests)
}

// applyLayout hands this frame's layout requests to the engine and adopts the
// patched display. A failure keeps the display the page already draws and is
// recorded in the motion stats, so one bad override never leaves a half-laid-out
// frame on screen.
func (p *Page) applyLayout(ctx context.Context, requests []motion.LayoutRequest) {
	applied := 0
	for i := range requests {
		if err := p.applyLayoutRequest(ctx, requests[i]); err != nil {
			p.motion.NoteLayout(applied, err)
			// The requests this frame did not apply go back to the head of the
			// queue. A geometry run's final destination is then retried on the
			// next frame instead of being pruned with the entry.
			p.motion.ReturnLayout(requests[i:])

			return
		}
		applied++
	}

	p.motion.NoteLayout(applied, nil)
}

// applyLayoutRequest relayouts one element with its sampled properties and
// adopts the result. The override values are the sampled CSS values, so the
// engine resolves percentages and relative units against the same containing
// block the static page used, and min/max clamping stays the engine's job.
func (p *Page) applyLayoutRequest(ctx context.Context, request motion.LayoutRequest) error {
	if request.ElementID == "" || len(request.Overrides) == 0 {
		return nil
	}

	overrides := make([]layout.Override, len(request.Overrides))
	for i := range request.Overrides {
		overrides[i] = layout.Override{
			ElementID: request.ElementID,
			Property:  request.Overrides[i].Property,
			Value:     request.Overrides[i].Value,
		}
	}

	patch, err := layout.RelayoutElementWithOverrides(
		ctx,
		p.cache.Styled(),
		p.draw.Display,
		request.ElementID,
		overrides,
		layout.Options{Images: p.renderState().Images, RetainState: true},
	)
	if err != nil {
		return fmt.Errorf("motion: relayout %q: %w", request.ElementID, err)
	}

	p.adoptLayout(patch)

	return nil
}

// adoptLayout installs the relaid-out display and rebinds every player. The
// rebind keeps each player's clock, so the element that just relaid out does
// not restart its own animation and its neighbours keep their phases. Damage
// comes from the display diff inside adoptSpliced: a width change moves the
// siblings outside the patched element's own box.
func (p *Page) adoptLayout(patch *layout.ElementPatch) bool {
	d := patch.Display
	if d == nil || d == p.draw.Display {
		return false
	}

	previous := p.adoptSpliced(d)
	p.layoutAdopting = true
	p.bindMotion(previous, d)
	p.layoutAdopting = false
	p.adoptTransitions(d)

	return true
}

// adoptTransitions points the live transition registries at the display a
// relayout patch produced. The patch already carries the sampled values, so the
// paint, layout-length, and discrete runs keep their clocks and endpoints; only
// the operations, boxes, and paint order they write move.
func (p *Page) adoptTransitions(d *Display) {
	if p.motion == nil {
		return
	}
	if p.transitionsOff {
		p.motion.ReleaseTransitions()

		return
	}
	p.motion.AdoptTransitions(d)
}

// setMotionView reports the page scroller state a scroll or view timeline
// samples against. The offset is the one the window last reported, so the
// animation follows the same geometry the picture is drawn at.
func (p *Page) setMotionView() {
	if p.motion == nil {
		return
	}

	x, y := p.scroll.Offset()
	content := p.contentRect()
	p.motion.SetView(motion.View{
		ScrollX:       float64(x),
		ScrollY:       float64(y),
		Width:         float64(p.width),
		Height:        float64(p.height),
		ContentWidth:  float64(content.Max.X),
		ContentHeight: float64(content.Max.Y),
	})
}

// restampMotion applies the 3D and perspective re-projection requests the last
// Step collected, at most one engine pass per tick.
//
// The pass needs the retained box tree a display only carries when it was laid
// out with RetainState. A page that animated a 3D transform or a perspective
// without it enables the state here and owes one redraw, so the first frame may
// not move; the next tick restamps. A 2D-only page collects no requests and
// never reaches the pass.
func (p *Page) restampMotion(ctx context.Context) {
	if p.motion == nil || p.animationsOff {
		return
	}

	requests := p.motion.TakeRestamp()
	if len(requests) == 0 {
		return
	}

	if !p.retainState || p.restampRedraw {
		p.retainState = true
		p.restampRedraw = true

		return
	}

	p.applyRestamp(ctx, requests)
}

// applyRestamp hands this tick's overrides to the engine and adopts the patched
// display. A failure keeps the display the page already draws and is recorded in
// the motion stats, so one bad override never leaves a half-projected frame on
// screen.
func (p *Page) applyRestamp(ctx context.Context, requests []motion.Request) {
	if p.draw.Display == nil || p.cache == nil {
		return
	}

	overrides := make([]layout.TransformOverride, len(requests))
	for i := range requests {
		overrides[i] = restampOverride(requests[i])
	}

	patch, err := layout.RestampTransforms(ctx, p.cache.Styled(), p.draw.Display, overrides,
		layout.Options{RetainState: true})
	if err != nil {
		p.motion.NoteRestamp(0, err)

		return
	}

	if !p.adoptRestamp(patch) {
		// The engine found no effective change, so it handed the page back.
		p.motion.NoteRestamp(0, nil)

		return
	}

	p.motion.NoteRestamp(len(requests), nil)
}

// sameOpSlice reports whether two displays share the same operation storage.
// A restamp mutates that storage in place, so the bound players stay valid and
// do not need to be rebuilt.
func sameOpSlice(previous, next *Display) bool {
	if previous == nil || next == nil || len(previous.Ops) != len(next.Ops) {
		return false
	}
	if len(previous.Ops) == 0 {
		return true
	}

	return &previous.Ops[0] == &next.Ops[0]
}

// restampOverride maps one sampled request onto the engine override. A
// perspective-only request leaves HasTransform false so the pass keeps the
// element's static transform; it replaces only the camera the element gives its
// descendants. An undeclared perspective-origin likewise leaves the resolved
// default in place.
func restampOverride(request motion.Request) layout.TransformOverride {
	return layout.TransformOverride{
		ElementID:            request.ElementID,
		Transform:            request.Transform,
		HasTransform:         request.Transform != "",
		Perspective:          request.Perspective,
		HasPerspective:       request.HasPerspective,
		PerspectiveOriginX:   request.PerspectiveOriginX,
		PerspectiveOriginY:   request.PerspectiveOriginY,
		HasPerspectiveOrigin: request.HasPerspectiveOrigin,
	}
}

// adoptRestamp installs the re-projected display and reports whether the
// engine changed anything. The generation bump tells the window the retained
// list changed, the rebuilt hit list keeps clicks on the projected boxes, and
// the rebind points every player at the new operations.
//
// Damage is the engine's Old and New painted area, marked once each. The restamp
// does not re-place anything, so the two are the same rectangle unless the
// projection moved what the subtree paints; marking the pair once is what keeps
// a 3D frame from dirtying the page twice through the display diff.
func (p *Page) adoptRestamp(patch *layout.ElementPatch) bool {
	d := patch.Display
	if d == nil || d == p.draw.Display {
		return false
	}

	previous := p.draw.SwapList(d)
	p.setBoxes(d.Boxes)
	p.hits = hitList(d.Boxes, fixedBoxes(d))
	p.markRect(patch.Old)
	if patch.New != patch.Old {
		p.markRect(patch.New)
	}
	p.generation++
	if sameOpSlice(previous, d) {
		p.motion.AdoptDisplay(d)
	} else {
		p.bindMotion(previous, d)
	}
	p.bindTransitions(nil)
	p.stats.repaints++

	regions := 1
	if patch.New != patch.Old {
		regions++
	}
	p.lastPatch = DisplayPatch{Changed: len(patch.Ops), Regions: regions}

	return true
}

// SetAnimations turns the automatic CSS motion driver on or off. It is on by
// default. Turning it off releases the bound players at once, so none writes
// through an operation the next publish moves; turning it on lets the next
// publish or tick bind again.
func (p *Page) SetAnimations(on bool) {
	p.animationsOff = !on
	if !on && p.motion != nil {
		_ = p.motion.Bind(nil)
	}
}

// Animations reports whether the automatic CSS motion driver is on.
func (p *Page) Animations() bool { return !p.animationsOff }

// SetTransitions turns the automatic CSS transition driver on or off. It is on
// by default. Turning it off releases the live renderers without writing, so a
// transition in flight freezes at the paint it had already presented instead of
// snapping to its base or its destination. The animation switch is separate;
// Page.Window() exposes this switch as SetTransitions.
func (p *Page) SetTransitions(on bool) {
	p.transitionsOff = !on
	if on || p.motion == nil {
		return
	}
	p.motion.ReleaseTransitions()
	if !p.motion.Active() {
		// A deadline the released renderer asked for would wake the window
		// once more for work that no longer exists.
		p.WantFrameAt(time.Time{})
	}
}

// Transitions reports whether the automatic CSS transition driver is on.
func (p *Page) Transitions() bool { return !p.transitionsOff }

// Motion returns the automatic CSS motion counters.
func (p *Page) Motion() MotionStats { return p.motion.Stats() }

// triggerPressState records the box whose animation-trigger entry actions
// fired on a pointer press. The click that follows the same gesture must not
// fire them a second time, and the release fires the exit actions for the same
// box even when the pointer moved before it lifted.
type triggerPressState struct {
	armed bool
	box   Box
}

// activationBox resolves the innermost box under a point the way the click
// sequence does: the box the point hit, or its innermost id-bearing ancestor.
func (p *Page) activationBox(x, y float64) (Box, bool) {
	box, ok := p.hitBox(x, y)
	if !ok {
		return Box{}, false
	}
	if box.ID == "" {
		if id := p.boxIDAt(x, y); id != "" {
			box = p.boxByID(id)
		}
	}

	return box, true
}

// pressTriggers fires the activated element's animation-trigger entry actions
// and remembers the gesture, so the click that follows does not fire them
// again and the release can fire the exit actions.
func (p *Page) pressTriggers(ctx context.Context, box Box) error {
	fired, err := p.fireTriggers(ctx, box, true)
	if err != nil {
		return err
	}
	if fired {
		p.triggerPress = triggerPressState{armed: true, box: box}
	}

	return nil
}

// releaseTriggers fires the exit actions of the box the press activated.
func (p *Page) releaseTriggers(ctx context.Context) error {
	state := p.triggerPress
	p.triggerPress = triggerPressState{}
	if !state.armed {
		return nil
	}
	_, err := p.fireTriggers(ctx, state.box, false)

	return err
}

// fireTriggers dispatches every animation-trigger name the activated element
// and its ancestors declare, at the event source the activation landed on. A
// click on a descendant activates the ancestor that owns the trigger, and a
// trigger-scope on either side still decides which listeners hear the event.
// It reports whether any name was dispatched.
//
// Trigger state is per effect, so a click reaches the same player the paint
// comes from, with the trigger scope the stylesheet declared. The page does
// not expose a programmatic trigger: an app that wants one changes the
// element's class or attribute, which is CSS.
func (p *Page) fireTriggers(ctx context.Context, box Box, active bool) (bool, error) {
	if err := useContext(ctx); err != nil {
		return false, err
	}
	if p.motion == nil || p.animationsOff {
		return false, nil
	}
	d := p.draw.Display
	if d == nil {
		return false, nil
	}
	index := boxIndexIn(d.Boxes, box)
	if index < 0 {
		return false, nil
	}
	names := triggerNamesAt(d.Boxes, index)
	if len(names) == 0 {
		return false, nil
	}
	passes := p.motion.TriggerPasses()
	if len(passes) == 0 {
		return false, nil
	}

	now := time.Now()
	for i := range passes {
		for _, name := range names {
			event := csstrigger.Event{Name: name, Source: index, Active: active}
			if err := csstrigger.Dispatch(passes[i].Boxes, event, passes[i].Players, now); err != nil {
				return true, err
			}
		}
	}
	// The sampled pose changes on the next tick, which is the same update for
	// a pointer event and the next frame for a programmatic click. A paused
	// player parks after one sample, so wake it to write the trigger's pose.
	p.motion.Wake()
	p.WantFrameIn(0)

	return true, nil
}

// triggerNamesAt collects the trigger names one activation raises: the
// dashed-ident animation-trigger names and the trigger-scope names declared by
// the activated element and its ancestors, nearest first. A card that scopes a
// name on its root therefore fires that name when any control inside it is
// activated, which is the catalog's Start-button pattern. Scope still decides
// which listeners respond.
func triggerNamesAt(boxes []Box, index int) []string {
	var names []string
	seen := make(map[string]bool)
	add := func(candidates []string) {
		for _, name := range candidates {
			if seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}

	for depth := 0; index >= 0 && index < len(boxes) && depth <= len(boxes); depth++ {
		for i := range boxes[index].Animations {
			add(triggerNames(boxes[index].Animations[i].Trigger))
		}
		add(triggerNames(boxes[index].TriggerScope))
		index = boxes[index].ParentIndex
	}

	return names
}

// triggerNames splits one animation-trigger declaration into its dashed-ident
// event names. Action keywords are skipped; the action mapping stays in
// csstrigger.
func triggerNames(value string) []string {
	var names []string
	for _, token := range strings.Fields(strings.ReplaceAll(value, ",", " ")) {
		if strings.HasPrefix(token, "--") {
			names = append(names, token)
		}
	}

	return names
}

// boxIndexIn returns the index of box in boxes, or -1. The hit list and the
// display carry copies of the same box data, so identity is the id, the parent
// link, and the rectangle.
func boxIndexIn(boxes []Box, box Box) int {
	for i := range boxes {
		if sameBox(boxes[i], box) {
			return i
		}
	}

	return -1
}

// sameBox compares the fields a box keeps across the hit list and the display.
func sameBox(a, b Box) bool {
	return a.ID == b.ID && a.Tag == b.Tag && a.ParentIndex == b.ParentIndex &&
		a.X == b.X && a.Y == b.Y && a.W == b.W && a.H == b.H
}

// activationTarget reports whether a focused control takes keyboard activation
// for its triggers: an enabled, non-editable control such as a button.
func (p *Page) activationTarget(field Control) bool {
	return field.ID != "" && !field.Disabled && !canEdit(field)
}

// activationKey reports whether key activates the focused control, the way a
// button responds to Enter and Space.
func activationKey(key string) bool {
	return key == "enter" || key == "space"
}
