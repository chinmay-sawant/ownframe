package page

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const transitionWireHTML = `<style>
body { margin: 0 }
#target { width: 40px; height: 20px; background-color: rgb(0,0,0); opacity: 1;
	transform: translateX(0px);
	transition-property: background-color, opacity, transform;
	transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { background-color: rgb(255,255,255); opacity: 0.5;
	transform: translateX(20px); }
</style><div id="target"></div>`

// newTransitionWirePage draws the class-change fixture with retained state so
// PatchElement can splice it.
func newTransitionWirePage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{HTML: transitionWireHTML, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// transitionWireOp returns the target element's first paint operation.
func transitionWireOp(t *testing.T, p *Page) *DisplayOp {
	t.Helper()

	box := p.boxByID("target")
	if box.ID == "" {
		t.Fatal("the target box is missing")
	}
	for _, index := range cssmotion.OperationIndices(p.Display(), box) {
		return &p.Display().Ops[index]
	}
	t.Fatal("the target box owns no operation")

	return nil
}

// patchTargetClass applies the fixture's class change through the element
// patch path, the way a click handler changes a class.
func patchTargetClass(t *testing.T, p *Page, class string) {
	t.Helper()
	mustPatch(t, p.PatchElement(context.Background(), "target",
		ElementEdit{Attribute: "class", Value: class}))
}

// transitionWireNear compares one paint field.
func transitionWireNear(got, want float64) bool { return math.Abs(got-want) <= 1e-3 }

// TestPatchElementStartsAutomaticTransitions is the phase gate at page level:
// one class change starts a colour, opacity, and transform transition without
// a single application call, and the sampled frame reaches the new paint.
func TestPatchElementStartsAutomaticTransitions(t *testing.T) {
	p := newTransitionWirePage(t)

	patchTargetClass(t, p, "on")
	wantMatrix := transitionWireOp(t, p).Transform()
	if got := p.Motion().TransitionsBound; got != 3 {
		t.Fatalf("bound transitions = %d, want 3", got)
	}

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.NextFrameAt().IsZero() {
		t.Fatal("a running transition did not ask for a frame")
	}
	op := transitionWireOp(t, p)
	if !transitionWireNear(op.R, 0) {
		t.Fatalf("first sample red = %v, want the from paint 0", op.R)
	}

	// Sample past the end with the page's own clock, so the test does not
	// depend on how long the scheduler takes to run the next tick.
	p.motion.Step(p.draw.Display, time.Now().Add(500*time.Millisecond))
	op = transitionWireOp(t, p)
	if !transitionWireNear(op.R, 1) {
		t.Fatalf("final red = %v, want 1", op.R)
	}
	if got := op.Opacity(); !transitionWireNear(got, .5) {
		t.Fatalf("final opacity = %v, want .5", got)
	}
	if got := op.Transform(); got != wantMatrix {
		t.Fatalf("final matrix = %+v, want the committed %+v", got, wantMatrix)
	}
	if got := p.Motion().TransitionsActive; got != 0 {
		t.Fatalf("finished transitions still active: %d", got)
	}
	p.NextFrameAt()
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a finished transition asked for a frame at %v", got)
	}
}

// TestInterruptedClassChangeReversesFromTheScreen checks the same rule through
// the page hook: the patch captures the operation the running renderer wrote,
// so the second run starts at the paint on screen.
func TestInterruptedClassChangeReversesFromTheScreen(t *testing.T) {
	p := newTransitionWirePage(t)
	ctx := context.Background()
	patchTargetClass(t, p, "on")
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	p.motion.Step(p.draw.Display, time.Now().Add(100*time.Millisecond))
	presented := transitionWireOp(t, p).R
	if presented <= 0 || presented >= 1 {
		t.Fatalf("the first run did not reach an intermediate paint: %v", presented)
	}

	patchTargetClass(t, p, "")
	// The reversal is shortened by the progress already spent, so the first
	// sample after the bind may already have advanced a little. It must still
	// start from the paint on screen, not from the run's base or destination.
	p.motion.Step(p.draw.Display, time.Now())
	if got := transitionWireOp(t, p).R; got > presented || got < presented*.5 {
		t.Fatalf("reversal start = %v, want the presented %v", got, presented)
	}

	// The reverse is shortened by the progress already spent, so it finishes
	// well inside the next 200ms rather than after the declared duration.
	p.motion.Step(p.draw.Display, time.Now().Add(200*time.Millisecond))
	if got := transitionWireOp(t, p).R; !transitionWireNear(got, 0) {
		t.Fatalf("shortened reversal end = %v, want 0", got)
	}
}

// TestTransitionsDisabledDoesNotBind checks the internal switch: a page with
// the transition driver off commits the new paint directly.
func TestTransitionsDisabledDoesNotBind(t *testing.T) {
	p := newTransitionWirePage(t)
	p.SetTransitions(false)
	if p.Transitions() {
		t.Fatal("SetTransitions(false) left the driver on")
	}

	patchTargetClass(t, p, "on")
	if got := p.Motion().TransitionsBound; got != 0 {
		t.Fatalf("bound transitions = %d, want 0", got)
	}
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	op := transitionWireOp(t, p)
	if !transitionWireNear(op.R, 1) || !transitionWireNear(op.Opacity(), .5) {
		t.Fatalf("disabled driver painted colour %v opacity %v, want the committed paint",
			op.R, op.Opacity())
	}
	if !p.NextFrameAt().IsZero() {
		t.Fatal("disabled driver asked for a frame")
	}
}

// TestTransitionsOffMidFlightFreezesThePresentedPaint checks that the switch
// drops the renderer without writing the base or the destination.
func TestTransitionsOffMidFlightFreezesThePresentedPaint(t *testing.T) {
	p := newTransitionWirePage(t)
	ctx := context.Background()
	patchTargetClass(t, p, "on")
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	p.motion.Step(p.draw.Display, time.Now().Add(100*time.Millisecond))
	presented := transitionWireOp(t, p).R

	p.SetTransitions(false)
	if got := p.Motion().TransitionsBound; got != 0 {
		t.Fatalf("live transitions after the switch = %d, want 0", got)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a released transition left a frame deadline at %v", got)
	}
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := transitionWireOp(t, p).R; !transitionWireNear(got, presented) {
		t.Fatalf("frozen paint = %v, want the presented %v", got, presented)
	}
}
