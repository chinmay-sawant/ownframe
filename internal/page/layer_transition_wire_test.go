package page

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// layerTransitionHTML wraps one stylesheet in the layer fixture page. The page
// keeps retained state so PatchElement can splice the class change.
func layerTransitionHTML(t *testing.T, css string) *Page {
	t.Helper()

	source := `<style>body { margin: 0 }` + css + `</style><div id="target"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// layerTransitionOp returns the target element's first paint operation.
func layerTransitionOp(t *testing.T, p *Page) *DisplayOp {
	t.Helper()

	for _, index := range targetOpIndices(p) {
		return &p.Display().Ops[index]
	}
	t.Fatal("the target box owns no operation")

	return nil
}

// targetOpIndices resolves the target box's operations from the display.
func targetOpIndices(p *Page) []int {
	box := p.boxByID("target")
	if box.ID == "" {
		return nil
	}
	var out []int
	for i := range p.Display().Ops {
		for _, id := range box.OperationIDs {
			if p.Display().Ops[i].ID == id {
				out = append(out, i)

				break
			}
		}
	}

	return out
}

// advanceTransitions samples the page's transitions and animations at a fixed
// offset from now, so a test does not depend on the scheduler.
func advanceTransitions(t *testing.T, p *Page, after time.Duration) {
	t.Helper()

	p.setMotionView()
	if p.animationsOff {
		p.motion.StepTransitions(p.draw.Display, time.Now().Add(after))

		return
	}
	p.motion.Step(p.draw.Display, time.Now().Add(after))
}

// transitionNumberNear compares two sampled numbers with room for the time the
// test spent between the bind and the sample.
func transitionNumberNear(got, want float64) bool { return math.Abs(got-want) <= .1 }

// TestLayerSpecsPromotesTransformTransition pins the page seam for a transform
// transition: one specification, the element's box, the sampled matrix, and the
// inverse base the renderer bakes against.
func TestLayerSpecsPromotesTransformTransition(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); transform: translateX(0px);
	transition-property: transform; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { transform: translateX(20px); }`)
	patchTargetClass(t, p, "on")
	advanceTransitions(t, p, 200*time.Millisecond)

	d := p.Display()
	specs := p.LayerSpecs(d)
	if len(specs) != 1 {
		t.Fatalf("specs = %+v, want one transform layer", specs)
	}
	spec := specs[0]
	op := layerTransitionOp(t, p)
	if !op.XformSet || op.Transform().IsIdentity() {
		t.Fatalf("sampled matrix = %+v, want a midpoint transform", op.Transform())
	}
	if !layerMatrixNear(spec.Matrix, op.Transform()) {
		t.Fatalf("matrix = %+v, want the sampled %+v", spec.Matrix, op.Transform())
	}
	if spec.Alpha != 1 {
		t.Fatalf("alpha = %v, want 1 for a transform transition", spec.Alpha)
	}
	if spec.Mode != "" {
		t.Fatalf("mode = %q, want normal", spec.Mode)
	}
	wantBase, ok := host.InvertLayerMatrix(spec.Matrix)
	if !ok {
		t.Fatalf("matrix %+v is not invertible", spec.Matrix)
	}
	if spec.Base != wantBase {
		t.Fatalf("base = %+v, want the inverse %+v", spec.Base, wantBase)
	}
}

// TestLayerSpecsPromotesOpacityTransition pins group alpha at the page seam: the
// specification carries the sampled opacity once, with an identity matrix, and
// the transition's overlap fades once because the renderer divides this value
// out of the bake.
func TestLayerSpecsPromotesOpacityTransition(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); opacity: 1;
	transition-property: opacity; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { opacity: 0.5; }`)
	patchTargetClass(t, p, "on")
	advanceTransitions(t, p, 200*time.Millisecond)

	specs := p.LayerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("specs = %+v, want one opacity layer", specs)
	}
	spec := specs[0]
	if !spec.Matrix.IsIdentity() {
		t.Fatalf("matrix = %+v, want the identity", spec.Matrix)
	}
	if spec.Base != layout.IdentityMatrix() {
		t.Fatalf("base = %+v, want the identity", spec.Base)
	}
	if !transitionNumberNear(spec.Alpha, .75) {
		t.Fatalf("alpha = %v, want the sampled .75", spec.Alpha)
	}
	if got := opOpacity(layerTransitionOp(t, p)); math.Abs(got-spec.Alpha) > 1e-6 {
		t.Fatalf("spec alpha = %v, want the operation's %v", spec.Alpha, got)
	}
	for _, index := range targetOpIndices(p) {
		op := &p.Display().Ops[index]
		if got := opOpacity(op); math.Abs(got-spec.Alpha) > 1e-6 {
			t.Fatalf("op %d alpha = %v, want the group %v", index, got, spec.Alpha)
		}
	}
}

// opOpacity returns the element opacity alone, without folding in a source
// alpha.
func opOpacity(op *DisplayOp) float64 {
	opacity := op.Opacity()
	if op.Alpha > 0 && op.Alpha < 1 && opacity > 0 {
		opacity /= op.Alpha
	}

	return opacity
}

// TestLayerSpecsReleasesFinishedTransition pins the release rule at the page
// seam: once the run ends, the hook names no layer, so the renderer releases
// the texture and repaints the pose it leaves.
func TestLayerSpecsReleasesFinishedTransition(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); transform: translateX(0px);
	transition-property: transform; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { transform: translateX(20px); }`)
	patchTargetClass(t, p, "on")
	advanceTransitions(t, p, 500*time.Millisecond)

	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs after the end = %+v, want nil", specs)
	}
	if got := p.Motion().TransitionsActive; got != 0 {
		t.Fatalf("active transitions after the end = %d, want 0", got)
	}
}

// TestLayerSpecsPromotesTransitionWithAnimationsOff pins the independent
// switches: turning the animation driver off must not disable transition
// layers.
func TestLayerSpecsPromotesTransitionWithAnimationsOff(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); transform: translateX(0px);
	transition-property: transform; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { transform: translateX(20px); }`)
	p.SetAnimations(false)
	patchTargetClass(t, p, "on")
	advanceTransitions(t, p, 200*time.Millisecond)

	if specs := p.LayerSpecs(p.Display()); len(specs) != 1 {
		t.Fatalf("specs = %+v, want one transition layer with animations off", specs)
	}

	p.SetTransitions(false)
	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs with both drivers off = %+v, want nil", specs)
	}
}

// TestLayerSpecsSharesOneLayerForAnimationAndTransition pins the merge rule at
// the page seam: a transform animation and an opacity transition on the same
// element produce one specification with both the matrix and the group alpha.
func TestLayerSpecsSharesOneLayerForAnimationAndTransition(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
@keyframes drift {
	0% { transform: translateX(0px); }
	100% { transform: translateX(40px); }
}
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); opacity: 1;
	animation-name: drift; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both;
	transition-property: opacity; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { opacity: 0.5; }`)
	patchTargetClass(t, p, "on")
	advanceTransitions(t, p, 200*time.Millisecond)

	entries := p.motion.LayerEntriesFor(p.Display())
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one shared entry", entries)
	}
	if !entries[0].Has2DTransform || !entries[0].HasOpacity {
		t.Fatalf("entry fields = %+v, want the animation matrix and transition alpha", entries[0])
	}

	specs := p.LayerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("specs = %+v, want one shared layer, not one per track", specs)
	}
	spec := specs[0]
	if spec.Matrix.IsIdentity() {
		t.Fatal("the animation matrix did not reach the layer")
	}
	if !transitionNumberNear(spec.Alpha, .75) {
		t.Fatalf("alpha = %v, want the transition's sampled .75", spec.Alpha)
	}
	if spec.BoxIndex != entries[0].BoxIndex {
		t.Fatalf("box = %d, want %d", spec.BoxIndex, entries[0].BoxIndex)
	}
}

// TestLayerSpecsSkips3DTransformTransition pins the projection rule: the
// capture refuses a 3D transform transition, counts the reason, and the page
// never names a layer for it.
func TestLayerSpecsSkips3DTransformTransition(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30); transform: rotateY(0deg);
	transition-property: transform; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { transform: rotateY(60deg); }`)
	patchTargetClass(t, p, "on")

	if got := p.Motion().TransitionsBound; got != 0 {
		t.Fatalf("bound 3D transitions = %d, want 0", got)
	}
	if got := p.Motion().TransitionsSkipped; got != 1 {
		t.Fatalf("skipped = %d, want the 3D transform counted", got)
	}
	advanceTransitions(t, p, 200*time.Millisecond)
	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs = %+v, want none for a 3D transform transition", specs)
	}
}

// TestLayerSpecsDisplayCrossFadeStaysPending keeps the follow-up honest: a
// transition on display never produces a layer. The engine refuses a patch that
// would drop the element's box entirely, so the outgoing-element cross-fade is
// still a documented gap rather than a fade-out layer this phase invents.
func TestLayerSpecsDisplayCrossFadeStaysPending(t *testing.T) {
	t.Parallel()

	p := layerTransitionHTML(t, `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60); opacity: 1;
	transition-property: opacity, display; transition-duration: 400ms;
	transition-timing-function: linear; }
#target.on { display: none; opacity: 0.5; }`)

	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs = %+v, want none for a display transition", specs)
	}

	err := p.PatchElement(context.Background(), "target",
		ElementEdit{Attribute: "class", Value: "on"})
	if err == nil {
		t.Fatal("a display: none patch was accepted; its outgoing layer is not implemented")
	}
	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs after the refused patch = %+v, want none; the cross-fade is pending", specs)
	}
}
