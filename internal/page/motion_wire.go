package page

import (
	"context"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// MotionStats is one snapshot of the automatic CSS motion counters and the
// last bind error. Probe().Motion returns it; the page copies the runtime
// block instead of logging on the hot path.
type MotionStats = motion.Stats

// bindMotion resolves the CSS animations of the display the page just
// published. Every publish path calls it after the display is installed and
// the generation bumped, so the runtime never samples an operation list the
// page has replaced. A fallback page has no patchable list, so it passes nil.
//
// A bind that finds a 3D transform keyframe, a 3D element, or a perspective
// keyframe turns retained layout state on, because the re-projection pass
// cannot run without it, and owes one redraw of the display the pass will
// patch. The frame the flag lands on does not move; the next tick restamps. A
// later publish of a retained display pays that debt, so a page that drew
// itself again before the tick does not wait an extra frame.
func (p *Page) bindMotion(display *Display) {
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

	_ = p.motion.Bind(display)
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
}

// stepMotion samples the bound animations, marks their damage, and asks for
// another frame while a player still has work. It runs before the user tick
// callback, so the callback's snapshot already carries the sampled paint and
// its diff never covers an animation change.
func (p *Page) stepMotion() {
	if p.motion == nil || p.animationsOff {
		return
	}

	damage := p.motion.Step(p.draw.Display, time.Now())
	if damage.Overflow {
		p.markFull()
	} else {
		for i := 0; i < damage.N; i++ {
			p.markRect(damage.Rects[i])
		}
	}

	if p.motion.Active() {
		p.WantFrameIn(0)
	}
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

	p.draw.SwapList(d)
	p.setBoxes(d.Boxes)
	p.hits = hitList(d.Boxes, fixedBoxes(d))
	p.markRect(patch.Old)
	if patch.New != patch.Old {
		p.markRect(patch.New)
	}
	p.generation++
	p.bindMotion(d)
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

// Motion returns the automatic CSS motion counters.
func (p *Page) Motion() MotionStats { return p.motion.Stats() }
