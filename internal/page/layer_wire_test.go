package page

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

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
