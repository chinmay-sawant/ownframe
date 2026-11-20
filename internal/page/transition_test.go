package page

import (
	"context"
	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"image"
	"math"
	"reflect"
	"testing"
	"time"
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

// layerFadeFrames animates only the group opacity, the property the current
// replay bake still writes into the owned operations.
const layerFadeFrames = `
@keyframes layer-fade {
	0% { opacity: 1; }
	100% { opacity: 0.25; }
}`

// layerMoveFadeFrames animates a transform and the group opacity together, the
// shape that used to stay direct because the bake wrote the sampled opacity
// into the owned operations.
const layerMoveFadeFrames = `
@keyframes layer-drift {
	0% { opacity: 0.2; transform: translateX(0px); }
	100% { opacity: 0.9; transform: translateX(40px); }
}`

// newLayerFadePage draws one element whose only animation is opacity.
func newLayerFadePage(t *testing.T) *Page {
	t.Helper()

	source := `<style>` + layerFadeFrames + `
#fader { width: 40px; height: 20px; background-color: rgb(20,40,60);
	animation-name: layer-fade; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both; }
</style><div id="fader"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// newLayerMoveFadePage draws one element that moves and fades together. Its
// border gives the renderer a second owned operation, which its promotion
// predicate requires.
func newLayerMoveFadePage(t *testing.T) *Page {
	t.Helper()

	source := `<style>` + layerMoveFadeFrames + `
#mover { width: 40px; height: 20px; background-color: rgb(20,40,60);
	border: 2px solid rgb(10,20,30);
	animation-name: layer-drift; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both; }
</style><div id="mover"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// tickLayerPage samples a running fixture after a short delay, so the entry
// carries a non-identity delta.
func tickLayerPage(t *testing.T, p *Page) {
	t.Helper()

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// layerMatrixNear compares two matrices term by term.
func layerMatrixNear(a, b layout.Matrix2D) bool {
	const eps = 1e-9

	return math.Abs(a.A-b.A) <= eps && math.Abs(a.B-b.B) <= eps &&
		math.Abs(a.C-b.C) <= eps && math.Abs(a.D-b.D) <= eps &&
		math.Abs(a.E-b.E) <= eps && math.Abs(a.F-b.F) <= eps
}

// TestLayerSpecsReportsRunningTransform pins the record the window composites:
// the target's box index, its sampled delta, the inverse base pose, and a group
// alpha of one.
func TestLayerSpecsReportsRunningTransform(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	tickLayerPage(t, p)

	d := p.Display()
	specs := p.LayerSpecs(d)
	if len(specs) != 1 {
		t.Fatalf("specs = %d, want one", len(specs))
	}

	spec := specs[0]
	wantBox := -1
	for i := range d.Boxes {
		if d.Boxes[i].ID == "target" {
			wantBox = i

			break
		}
	}
	if spec.ElementID != "target" || spec.BoxIndex != wantBox {
		t.Fatalf("spec = %q at %d, want target at %d", spec.ElementID, spec.BoxIndex, wantBox)
	}

	op := &d.Ops[motionTargetIndex(t, p)]
	if !op.XformSet {
		t.Fatal("the sampler left no matrix on the target operation")
	}
	if !layerMatrixNear(spec.Matrix, op.Transform()) {
		t.Fatalf("matrix = %+v, want the sampled %+v", spec.Matrix, op.Transform())
	}
	if spec.Matrix.IsIdentity() {
		t.Fatal("the fixture sampled the identity matrix")
	}
	if spec.Alpha != 1 {
		t.Fatalf("alpha = %v, want 1 for a transform-only entry", spec.Alpha)
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
	if got := spec.Matrix.Mul(spec.Base); !layerMatrixNear(got, layout.IdentityMatrix()) {
		t.Fatalf("matrix * base = %+v, want the identity", got)
	}
}

// TestLayerSpecsStaticPageReturnsNil keeps a page with no animation on the
// direct replay path.
func TestLayerSpecsStaticPageReturnsNil(t *testing.T) {
	t.Parallel()

	p, err := New(Config{
		HTML:   `<body style="margin:0"><div id="bar" style="width:40px;height:12px;background:#224466"></div></body>`,
		Width:  80,
		Height: 40,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs = %+v, want nil for a static page", specs)
	}
}

// TestLayerSpecsAnimationsOffReturnsNil keeps the page-wide off switch ahead of
// the motion accessor.
func TestLayerSpecsAnimationsOffReturnsNil(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	tickLayerPage(t, p)
	if got := len(p.LayerSpecs(p.Display())); got != 1 {
		t.Fatalf("specs with the driver on = %d, want one", got)
	}

	p.SetAnimations(false)
	if specs := p.LayerSpecs(p.Display()); specs != nil {
		t.Fatalf("specs = %+v, want nil with animations off", specs)
	}
}

// TestLayerSpecsStaleDisplayReturnsNil refuses a display the page replaced, so
// the renderer never captures a box tree the window is not drawing.
func TestLayerSpecsStaleDisplayReturnsNil(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	tickLayerPage(t, p)

	if specs := p.LayerSpecs(&layout.Display{}); specs != nil {
		t.Fatalf("specs = %+v, want nil for a stale display", specs)
	}
	if specs := p.LayerSpecs(nil); specs != nil {
		t.Fatalf("specs = %+v, want nil without a display", specs)
	}
}

// TestLayerSpecsPromotesOpacityOnly pins that a pure opacity entry reaches the
// renderer. The bake divides the sampled opacity out and the composite applies
// it once, so an identity matrix no longer keeps the element direct.
func TestLayerSpecsPromotesOpacityOnly(t *testing.T) {
	t.Parallel()

	p := newLayerFadePage(t)
	tickLayerPage(t, p)

	entries := p.motion.LayerEntries()
	if len(entries) != 1 || !entries[0].HasOpacity || entries[0].Has2DTransform {
		t.Fatalf("entries = %+v, want one opacity entry", entries)
	}

	specs := p.LayerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("specs = %+v, want one opacity layer", specs)
	}
	spec := specs[0]
	if !spec.Matrix.IsIdentity() {
		t.Fatalf("matrix = %+v, want the identity for an opacity-only entry", spec.Matrix)
	}
	if spec.Base != layout.IdentityMatrix() {
		t.Fatalf("base = %+v, want the identity", spec.Base)
	}
	if spec.Alpha <= 0 || spec.Alpha >= 1 {
		t.Fatalf("alpha = %v, want a sampled opacity between 0 and 1", spec.Alpha)
	}
	if spec.Alpha != entries[0].Alpha {
		t.Fatalf("alpha = %v, want the entry's %v", spec.Alpha, entries[0].Alpha)
	}
	if spec.BoxIndex != entries[0].BoxIndex {
		t.Fatalf("box = %d, want %d", spec.BoxIndex, entries[0].BoxIndex)
	}
}

// TestLayerSpecsPromotesTransformAndOpacity pins the combined entry: the matrix
// reaches the renderer as it is, and the group alpha rides beside it instead of
// keeping the element direct.
func TestLayerSpecsPromotesTransformAndOpacity(t *testing.T) {
	t.Parallel()

	p := newLayerMoveFadePage(t)
	tickLayerPage(t, p)

	entries := p.motion.LayerEntries()
	if len(entries) != 1 || !entries[0].Has2DTransform || !entries[0].HasOpacity {
		t.Fatalf("entries = %+v, want one transform and opacity entry", entries)
	}

	specs := p.LayerSpecs(p.Display())
	if len(specs) != 1 {
		t.Fatalf("specs = %+v, want one promoted layer", specs)
	}
	spec := specs[0]
	if spec.Matrix.IsIdentity() {
		t.Fatal("the fixture sampled the identity matrix")
	}
	if !layerMatrixNear(spec.Matrix, entries[0].Matrix) {
		t.Fatalf("matrix = %+v, want the entry's %+v", spec.Matrix, entries[0].Matrix)
	}
	if spec.Alpha <= 0 || spec.Alpha >= 1 {
		t.Fatalf("alpha = %v, want the sampled group opacity", spec.Alpha)
	}
	wantBase, ok := host.InvertLayerMatrix(spec.Matrix)
	if !ok {
		t.Fatalf("matrix %+v is not invertible", spec.Matrix)
	}
	if spec.Base != wantBase {
		t.Fatalf("base = %+v, want the inverse %+v", spec.Base, wantBase)
	}
	if got := spec.Matrix.Mul(spec.Base); !layerMatrixNear(got, layout.IdentityMatrix()) {
		t.Fatalf("matrix * base = %+v, want the identity", got)
	}
}

// TestLayerSpecsSkipsProjection covers the 3D and perspective samples the
// accessor refuses, so no projective map reaches a layer matrix.
func TestLayerSpecsSkipsProjection(t *testing.T) {
	t.Parallel()

	t.Run("3d keyframe", func(t *testing.T) {
		p := newRestampPage(t, true)
		tickLayerPage(t, p)

		if entries := p.motion.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a 3D sample", entries)
		}
		if specs := p.LayerSpecs(p.Display()); specs != nil {
			t.Fatalf("specs = %+v, want nil for a 3D sample", specs)
		}
	})

	t.Run("perspective keyframe", func(t *testing.T) {
		p := newPerspectiveCameraPage(t, true)
		tickLayerPage(t, p)

		if entries := p.motion.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a perspective sample", entries)
		}
		if specs := p.LayerSpecs(p.Display()); specs != nil {
			t.Fatalf("specs = %+v, want nil for a perspective sample", specs)
		}
	})
}

// TestLayerSpecsNilPageReturnsNil keeps the role safe before a page is built.
func TestLayerSpecsNilPageReturnsNil(t *testing.T) {
	t.Parallel()

	var p *Page
	if specs := p.LayerSpecs(nil); specs != nil {
		t.Fatalf("specs = %+v, want nil for a nil page", specs)
	}
}

// hoverBoxWidth returns the CSS-pixel width of the box with the given id.
func hoverBoxWidth(t *testing.T, p *Page, id string) float64 {
	t.Helper()

	for _, b := range p.Boxes() {
		if b.ID == id {
			return b.W
		}
	}

	t.Fatalf("no box %q", id)

	return 0
}

// TestStyleTriggerHoverSkipsCascade proves a hover on a hoverless sheet
// skips the style stage: no parse, no cascade, no layout, and no repaint.
// Counters prove it, not timing.
func TestStyleTriggerHoverSkipsCascade(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#e{width:100px;height:20px}</style></head><body><div id="e">x</div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	before := p.Stats()
	if err := p.Hover(ctx, 10, 10); err != nil {
		t.Fatal(err)
	}

	if got := p.Stats(); !reflect.DeepEqual(got, before) {
		t.Fatalf("hover restyled a hoverless sheet: %+v -> %+v", before, got)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	after := p.Stats()
	if after.Parses != before.Parses || after.Cascades != before.Cascades {
		t.Fatalf("redraw recascaded a hoverless sheet: %+v -> %+v", before, after)
	}
	if after.Layouts != before.Layouts || after.Repaints != before.Repaints {
		t.Fatalf("redraw relaid a hoverless sheet: %+v -> %+v", before, after)
	}
}

// TestStyleTriggerStateNoopReusesDocument proves the style fast path: a
// redraw that only moves state the sheets cannot match reuses the cached
// styled document instead of re-cascading. A move on a used pseudo still
// takes the full relayout.
func TestStyleTriggerStateNoopReusesDocument(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#e{width:100px;height:20px}#e:hover{width:120px}</style></head><body><div id="e">x</div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.statePseudos.Hover {
		t.Fatal("sheets should use :hover")
	}

	cached := p.cache.Styled()
	if cached == nil {
		t.Fatal("no cached document")
	}

	// Focus is unused by these sheets, so moving it must not restyle.
	focused := p.renderState()
	focused.Focus = "e"
	doc, ok := p.styleNoop(focused)
	if !ok {
		t.Fatal("unused focus move was not a style noop")
	}
	if doc != cached {
		t.Fatal("unused focus move replaced the styled document")
	}
	if routed, err := p.styleStaged(ctx, false, p.source, focused); err != nil {
		t.Fatal(err)
	} else if routed != cached {
		t.Fatal("styleStaged re-cascaded an unused focus move")
	}

	// Hover is used, so moving it must restyle.
	hovered := p.renderState()
	hovered.Hover = "e"
	if _, ok := p.styleNoop(hovered); ok {
		t.Fatal("used hover move skipped the restyle")
	}
	if doc, err := p.styleStaged(ctx, false, p.source, hovered); err != nil {
		t.Fatal(err)
	} else if doc == cached {
		t.Fatal("used hover move reused the styled document")
	}

	// The noop leaves the cascade counters alone.
	if got := p.Stats(); got.Parses != 1 || got.Cascades != 1 {
		t.Fatalf("noop restyle moved the cascade counters: %+v", got)
	}
}

// TestStyleTriggerThemeSwapApplies proves a same-source theme swap paints
// the new sheet even when the style cache is warm: the swap takes a fresh
// Apply instead of the cacheFresh relayout, which would keep the old extra
// sheet. It also refreshes which pseudos the page reports.
func TestStyleTriggerThemeSwapApplies(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#swatch{width:100px;height:40px}</style></head><body><div id="swatch"></div></body></html>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := hoverBoxWidth(t, p, "swatch"); got != 100 {
		t.Fatalf("plain width = %v, want 100", got)
	}
	if p.statePseudos.Hover {
		t.Fatal("plain sheets should not use :hover")
	}

	// Swap the theme without dropping the parsed tree, the way a
	// tree-preserving SetTheme would. The next redraw must still paint it.
	sheet, err := css.Parse(`#swatch{width:150px}#swatch:hover{width:160px}`)
	if err != nil {
		t.Fatal(err)
	}
	p.theme = sheet
	p.themeSrc = `#swatch{width:150px}#swatch:hover{width:160px}`

	before := p.Stats()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := hoverBoxWidth(t, p, "swatch"); got != 150 {
		t.Fatalf("themed width = %v, want 150", got)
	}
	if !p.statePseudos.Hover {
		t.Fatal("theme :hover was not reported after the swap")
	}
	if got := p.Stats(); got.Cascades == before.Cascades {
		t.Fatalf("theme swap skipped the cascade: %+v", got)
	}
}

// TestStyleTriggerThemeParity proves the SetTheme path paints exactly what a
// full rebuild with the same themed document paints.
func TestStyleTriggerThemeParity(t *testing.T) {
	t.Parallel()

	source := `<html><head><style>#swatch{width:100px;height:40px}</style></head><body><div id="swatch"></div></body></html>`
	theme := `#swatch{width:150px;background:#00ff00}`

	ctx := context.Background()

	swapped, err := New(Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := swapped.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := swapped.SetTheme(theme); err != nil {
		t.Fatal(err)
	}
	if err := swapped.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	built, err := New(Config{HTML: source, Theme: theme, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := built.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got, want := hoverBoxWidth(t, swapped, "swatch"), hoverBoxWidth(t, built, "swatch"); got != want {
		t.Fatalf("swapped width = %v, rebuilt width = %v", got, want)
	}
}

// TestStyleTriggerKeyComparisons pins the redraw key behavior the style
// fast path relies on: only real input changes refuse the skip.
func TestStyleTriggerKeyComparisons(t *testing.T) {
	t.Parallel()

	view := draw.Viewport{Width: 320, Height: 200}
	state := render.State{Hover: "e"}
	pseudos := render.StatePseudos{Hover: true}

	last := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view, state, pseudos, 0)
	if !draw.Unchanged(true, last, last) {
		t.Fatal("identical keys report changed")
	}
	if draw.Unchanged(false, last, last) {
		t.Fatal("missing last frame reports unchanged")
	}

	same := last
	if !draw.Unchanged(true, last, same) {
		t.Fatal("equal keys report changed")
	}

	for name, mutate := range map[string]func(*draw.Key){
		"width":  func(k *draw.Key) { k.Width++ },
		"height": func(k *draw.Key) { k.Height++ },
		"hover":  func(k *draw.Key) { k.Hover = "other" },
		"focus":  func(k *draw.Key) { k.Focus = "f" },
		"active": func(k *draw.Key) { k.Active = "a" },
		"images": func(k *draw.Key) { k.Images++ },
		"rev":    func(k *draw.Key) { k.SourceRev++ },
		"hash":   func(k *draw.Key) { k.SourceHash++ },
		"len":    func(k *draw.Key) { k.SourceLen++ },
	} {
		next := last
		mutate(&next)
		if draw.Unchanged(true, last, next) {
			t.Fatalf("%s change reports unchanged", name)
		}
	}

	// Equal revision, hash, and length never memcmp the source: different
	// Source bytes with the same identity still report unchanged.
	collision := last
	collision.Source = "b"
	if len(collision.Source) != collision.SourceLen {
		t.Fatal("fixture sources must share one length")
	}
	collision.SourceHash = last.SourceHash
	if !draw.Unchanged(true, last, collision) {
		t.Fatal("revision-identical keys memcmp the source")
	}

	// Without revisions, different sources still report changed.
	a := draw.KeyFor("aaa", view, render.State{}, render.StatePseudos{}, 0)
	b := draw.KeyFor("bbb", view, render.State{}, render.StatePseudos{}, 0)
	if draw.Unchanged(true, a, b) {
		t.Fatal("different sources report unchanged")
	}
	if !draw.Unchanged(true, a, a) {
		t.Fatal("identical sourceless keys report changed")
	}

	// An unmatched state drops out of the key, so hovering a hoverless
	// sheet never changes the frame.
	bare := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view,
		render.State{Hover: "e"}, render.StatePseudos{}, 0)
	plain := draw.KeyForHashed("a", draw.HashString("a"), 1, 1, view,
		render.State{}, render.StatePseudos{}, 0)
	if !draw.Unchanged(true, bare, plain) {
		t.Fatal("unused hover state changes the key")
	}
}

// newTransitionLayoutPage draws a data-bound class fixture with retained state
// so a geometry transition can splice the display after a redraw.
func newTransitionLayoutPage(t *testing.T, source string, data any) *Page {
	t.Helper()

	p, err := New(Config{HTML: source, Width: 320, Height: 160, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// transitionClassData is the bound class on the fixture's target.
type transitionClassData struct {
	Class string
}

// redrawClass changes the bound class and publishes it.
func redrawClass(t *testing.T, p *Page, data *transitionClassData, class string) {
	t.Helper()
	data.Class = class
	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
}

// TestWidthTransitionWidensTheBoxAndMovesASibling is the page-level geometry
// proof: a class change starts a width transition, the relayout grows the box,
// the in-flow sibling moves with it, and the display diff damages the sibling.
func TestWidthTransitionWidensTheBoxAndMovesASibling(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#row { display: flex; width: 300px; height: 60px; }
#a { width: 40px; height: 20px; background-color: rgb(20,40,60);
	transition-property: width; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { width: 120px; }
#b { width: 40px; height: 20px; background-color: rgb(80,100,120); }
</style><div id="row"><div id="a" class="{{.Class}}"></div><div id="b"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	redrawClass(t, p, data, "on")
	start := time.Now()

	if got := p.Motion().TransitionsBound; got != 1 {
		t.Fatalf("bound transitions = %d, want the width run", got)
	}
	p.TakeDirty()
	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))

	a, b := p.boxByID("a"), p.boxByID("b")
	if a.W <= 40 || a.W >= 120 {
		t.Fatalf("midpoint width = %v, want between 40 and 120", a.W)
	}
	if b.X <= 40 {
		t.Fatalf("sibling X = %v, want it past the base 40", b.X)
	}
	if got := p.Motion().LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts = %d, want 1", got)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("relayout error = %q, want none", got)
	}

	want := image.Rect(int(b.X), int(b.Y), int(b.X+b.W), int(b.Y+b.H))
	rects, ok := p.TakeDirtyRects()
	if !ok {
		t.Fatal("the width transition marked no damage")
	}
	covered := false
	for _, rect := range rects {
		if rect.Min.X <= want.Min.X && rect.Min.Y <= want.Min.Y &&
			rect.Max.X >= want.Max.X && rect.Max.Y >= want.Max.Y {
			covered = true
		}
	}
	if !covered {
		t.Fatalf("damage %v does not cover the moved sibling %v", rects, want)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	if got := p.boxByID("a").W; got != 120 {
		t.Fatalf("final width = %v, want the destination 120", got)
	}
	if got := p.boxByID("b").X; got != 120 {
		t.Fatalf("final sibling X = %v, want 120 beside the final width", got)
	}
}

// TestHeightTransitionGrowsTheBox checks the vertical extent through the same
// relayout path. The fixed-height row keeps the change inside the patch.
func TestHeightTransitionGrowsTheBox(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#row { width: 200px; height: 120px; }
#a { height: 20px; background-color: rgb(20,40,60);
	transition-property: height; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { height: 80px; }
</style><div id="row"><div id="a" class="{{.Class}}"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	redrawClass(t, p, data, "on")
	start := time.Now()

	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))
	if got := p.boxByID("a").H; got <= 20 || got >= 80 {
		t.Fatalf("midpoint height = %v, want between 20 and 80", got)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("height relayout error = %q, want none", got)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	if got := p.boxByID("a").H; got != 80 {
		t.Fatalf("final height = %v, want 80", got)
	}
}

// TestLeftTopTransitionMovesTheBox checks the inset path: the position change
// becomes a relayout override and the element's border box follows it.
func TestLeftTopTransitionMovesTheBox(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#stage { position: relative; width: 300px; height: 150px; }
#a { position: absolute; left: 0; top: 0; width: 40px; height: 20px;
	background-color: rgb(20,40,60);
	transition-property: left, top; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { left: 100px; top: 60px; }
</style><div id="stage"><div id="a" class="{{.Class}}"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	redrawClass(t, p, data, "on")
	start := time.Now()

	if got := p.Motion().TransitionsBound; got != 2 {
		t.Fatalf("bound transitions = %d, want left and top", got)
	}
	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))
	box := p.boxByID("a")
	if box.X <= 0 || box.X >= 100 {
		t.Fatalf("midpoint left = %v, want between 0 and 100", box.X)
	}
	if box.Y <= 0 || box.Y >= 60 {
		t.Fatalf("midpoint top = %v, want between 0 and 60", box.Y)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("inset relayout error = %q, want none", got)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	box = p.boxByID("a")
	if box.X != 100 || box.Y != 60 {
		t.Fatalf("final box = (%v,%v), want (100,60)", box.X, box.Y)
	}
}

// TestVisibilityTransitionHidesTheRemainingFill checks the presence path at the
// page seam: a leaving element keeps the fill it still owns visible for the run
// and hides it at the end.
func TestVisibilityTransitionHidesTheRemainingFill(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	transition-property: visibility; transition-duration: 400ms;
	transition-behavior: allow-discrete; transition-timing-function: linear; }
#target.on { visibility: hidden; }
</style><div id="target" class="{{.Class}}">hidden words</div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	redrawClass(t, p, data, "on")
	start := time.Now()

	box := p.boxByID("target")
	if len(box.OperationIDs) == 0 {
		t.Fatal("the hidden element kept no operation to fade")
	}
	op := transitionWireOp(t, p)
	if op.Alpha != 1 {
		t.Fatalf("alpha at the bind = %v, want the visible 1", op.Alpha)
	}
	p.setMotionView()
	p.motion.Step(p.draw.Display, start.Add(200*time.Millisecond))
	op = transitionWireOp(t, p)
	if op.Alpha != 1 {
		t.Fatalf("alpha mid-run = %v, want the visible 1", op.Alpha)
	}

	p.motion.Step(p.draw.Display, start.Add(600*time.Millisecond))
	op = transitionWireOp(t, p)
	if op.Alpha > 0.01 {
		t.Fatalf("alpha at the end = %v, want the hidden endpoint; stats %+v",
			op.Alpha, p.Motion())
	}
}

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
	for _, index := range motion.OperationIndices(p.Display(), box) {
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
