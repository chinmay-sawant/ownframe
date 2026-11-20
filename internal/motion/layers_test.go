package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"math"
	"strings"
	"testing"
	"time"
)

// layerPairDisplay adds a second element that fades to animatedDisplay's
// translating chip, so one read exercises a transform entry and an opacity
// entry side by side. Both players loop, so both stay active after a sample.
func layerPairDisplay() *layout.Display {
	d := animatedDisplay("running", true)

	d.Ops = append(d.Ops, layout.DisplayOp{
		ID: testOpID + 1, Kind: layout.DisplayOpFillRect,
		X: 60, Y: 10, W: 20, H: 20, R: 0.2, G: 0.4, B: 0.6, Alpha: 1,
	})
	d.OpOwners = []int{0, 1}
	d.Boxes = append(d.Boxes, layout.Box{
		ID: "fader", Tag: "div", X: 60, Y: 10, W: 20, H: 20, ParentIndex: -1,
		OperationIDs: []uint64{testOpID + 1},
		Animations: []layout.Animation{{
			Name: testFade, DurationMS: testDurationMS, IterationInfinite: true,
			FillMode: "both",
		}},
	})
	d.Keyframes = append(d.Keyframes, layout.KeyframesRule{
		Name: testFade,
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "1"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "0.25"},
			}},
		},
	})

	return d
}

// layer3DDisplay turns the slide keyframes into a rotateY pair, which the
// sampler cannot express as a 2D matrix.
func layer3DDisplay() *layout.Display {
	d := animatedDisplay("running", true)
	for i := range d.Keyframes[0].Keyframes {
		d.Keyframes[0].Keyframes[i].Declarations[0].Value = "rotateY(45deg)"
	}

	return d
}

// rotatedDisplay replaces the slide keyframes with a 0deg..90deg rotation and
// stamps the same static 30deg transform on the box and its operation, the way
// the engine lays a rotated element out.
func rotatedDisplay() *layout.Display {
	d := animatedDisplay("running", true)
	for i := range d.Keyframes[0].Keyframes {
		d.Keyframes[0].Keyframes[i].Declarations[0].Value = "rotate(0deg)"
	}
	d.Keyframes[0].Keyframes[1].Declarations[0].Value = "rotate(90deg)"

	static := layerRotate(30)
	d.Ops[0].SetXform(static)
	d.Boxes[0].Transform = layout.TransformInfo{
		Parent: layout.IdentityMatrix(), World: static, HasMatrix: true,
	}

	return d
}

// translatedDisplay is the same shape with a static translation, so the delta
// carries a translation that the base pose already held.
func translatedDisplay() *layout.Display {
	d := animatedDisplay("running", true)
	static := layerTranslate(5, 0)
	d.Ops[0].SetXform(static)
	d.Boxes[0].Transform = layout.TransformInfo{
		Parent: layout.IdentityMatrix(), World: static, HasMatrix: true,
	}

	return d
}

// layerTranslate builds the display-space translation the engine stamps.
func layerTranslate(x, y float64) layout.Matrix2D {
	return layout.Matrix2D{A: 1, D: 1, E: x, F: y}
}

// layerRotate builds a CSS y-down rotation about the origin.
func layerRotate(deg float64) layout.Matrix2D {
	sin, cos := math.Sincos(deg * math.Pi / 180)

	return layout.Matrix2D{A: cos, B: sin, C: -sin, D: cos}
}

// scaledDisplay replaces the slide keyframes with a scale from zero, whose
// first sample is a singular delta a layer cannot invert.
func scaledDisplay() *layout.Display {
	d := animatedDisplay("running", true)
	d.Keyframes[0].Keyframes[0].Declarations[0].Value = "scale(0)"
	d.Keyframes[0].Keyframes[1].Declarations[0].Value = "scale(1)"

	return d
}

// layerFloatClose compares two sampled coordinates.
func layerFloatClose(got, want float64) bool {
	return math.Abs(got-want) <= 1e-9
}

// layerMatrixClose compares two matrices term by term.
func layerMatrixClose(got, want layout.Matrix2D) bool {
	return layerFloatClose(got.A, want.A) && layerFloatClose(got.B, want.B) &&
		layerFloatClose(got.C, want.C) && layerFloatClose(got.D, want.D) &&
		layerFloatClose(got.E, want.E) && layerFloatClose(got.F, want.F)
}

// TestLayerEntriesReportsTransformAndOpacity pins the record a page maps to a
// retained layer: two entries, the transform delta in display points, and the
// group alpha.
func TestLayerEntriesReportsTransformAndOpacity(t *testing.T) {
	r := New()
	r.now = testClock
	d := layerPairDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	entries := r.LayerEntries()
	if len(entries) != 2 {
		t.Fatalf("entries = %d, want two", len(entries))
	}

	mover := entries[0]
	if mover.ElementID != testChipID || mover.BoxIndex != 0 {
		t.Fatalf("mover = %q at %d, want %q at 0", mover.ElementID, mover.BoxIndex, testChipID)
	}
	if !mover.Has2DTransform || mover.HasOpacity {
		t.Fatalf("mover fields = %+v, want a transform entry only", mover)
	}
	if want := (layout.Matrix2D{A: 1, D: 1, E: 10}); !layerMatrixClose(mover.Matrix, want) {
		t.Fatalf("mover matrix = %+v, want the 10-point midpoint %+v", mover.Matrix, want)
	}
	if mover.Alpha != 1 {
		t.Fatalf("mover alpha = %v, want 1 for a transform-only entry", mover.Alpha)
	}
	if mover.Mode != LayerModeNormal {
		t.Fatalf("mover mode = %q, want normal", mover.Mode)
	}
	if mover.HasRestamp || mover.HasLayout || mover.Paused || !mover.Active {
		t.Fatalf("mover flags = %+v, want an active, unblocked entry", mover)
	}

	fader := entries[1]
	if fader.ElementID != "fader" || fader.BoxIndex != 1 {
		t.Fatalf("fader = %q at %d, want fader at 1", fader.ElementID, fader.BoxIndex)
	}
	if fader.Has2DTransform || !fader.HasOpacity {
		t.Fatalf("fader fields = %+v, want an opacity entry only", fader)
	}
	if !fader.Matrix.IsIdentity() {
		t.Fatalf("fader matrix = %+v, want the identity", fader.Matrix)
	}
	if want := 0.625; !layerFloatClose(fader.Alpha, want) {
		t.Fatalf("fader alpha = %v, want the 0.625 midpoint", fader.Alpha)
	}
}

// TestLayerEntriesAlphaIsGroupOpacity reads a sampled opacity through an
// operation whose own colour alpha is 0.5. The group alpha is 0.4375; the
// per-operation product is 0.21875, which would double-darken the layer.
func TestLayerEntriesAlphaIsGroupOpacity(t *testing.T) {
	r := New()
	r.now = testClock
	d := opacityDisplay()
	d.Ops[0].Alpha = 0.5

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(3*testDurationMS/4*time.Millisecond))

	entries := r.LayerEntries()
	if len(entries) != 1 {
		t.Fatalf("entries = %d, want one", len(entries))
	}
	if !entries[0].HasOpacity || entries[0].Has2DTransform {
		t.Fatalf("entry fields = %+v, want opacity only", entries[0])
	}
	if want := 0.4375; !layerFloatClose(entries[0].Alpha, want) {
		t.Fatalf("alpha = %v, want the group %v, not the per-op product", entries[0].Alpha, want)
	}
}

// TestLayerEntriesMatrixRecoversTheDelta checks the inverse rule against a
// static transform that the sampled operation matrix already holds. The layer
// delta must take the captured pose to the sampled pose, in display points.
func TestLayerEntriesMatrixRecoversTheDelta(t *testing.T) {
	t.Run("translation", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := translatedDisplay()

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		entries := r.LayerEntries()
		if len(entries) != 1 {
			t.Fatalf("entries = %d, want one", len(entries))
		}
		want := layerTranslate(5, 0)
		if !layerMatrixClose(entries[0].Matrix, want) {
			t.Fatalf("delta = %+v, want %+v", entries[0].Matrix, want)
		}
		// The delta composed with the captured pose is the sampled pose.
		current := d.Ops[0].Transform()
		if got := entries[0].Matrix.Mul(layerTranslate(5, 0)); !layerMatrixClose(got, current) {
			t.Fatalf("delta * base = %+v, want the sampled %+v", got, current)
		}
	})

	t.Run("rotation", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := rotatedDisplay()

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		entries := r.LayerEntries()
		if len(entries) != 1 {
			t.Fatalf("entries = %d, want one", len(entries))
		}
		// The sample replaces the static 30deg transform with 45deg, so the
		// page-space delta from the captured pose is 15deg.
		want := layerRotate(15)
		if !layerMatrixClose(entries[0].Matrix, want) {
			t.Fatalf("delta = %+v, want %+v", entries[0].Matrix, want)
		}
		current := d.Ops[0].Transform()
		if got := entries[0].Matrix.Mul(layerRotate(30)); !layerMatrixClose(got, current) {
			t.Fatalf("delta * base = %+v, want the sampled %+v", got, current)
		}
	})
}

// TestLayerEntriesFilters covers the conditions a layer cannot carry.
func TestLayerEntriesFilters(t *testing.T) {
	t.Run("colour entry", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := splitDisplay("running")

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none: the colour entry would re-bake the layer every frame", entries)
		}
	})

	t.Run("3d keyframe", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := layer3DDisplay()

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a 3D sample", entries)
		}
		if got := r.Stats().NeedsRestamp; got != 1 {
			t.Fatalf("restamp count = %d, want the 3D sample counted", got)
		}
	})

	t.Run("perspective keyframe", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := perspectiveDisplay()

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a perspective sample", entries)
		}
	})

	t.Run("3d element", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := animatedDisplay("running", true)
		d.Boxes[0].Transform.HasTransform3D = true

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a 3D element", entries)
		}
	})

	t.Run("paused", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := animatedDisplay("paused", true)

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

		if !r.entries[0].player.Paused() {
			t.Fatal("the fixture's player is not paused")
		}
		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a paused pose", entries)
		}
	})

	t.Run("layout field", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := animatedDisplay("running", true)

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))
		if got := len(r.LayerEntries()); got != 1 {
			t.Fatalf("entries before the field = %d, want one", got)
		}

		// No CSS source binds X/Y today: the legacy move path is the only
		// writer, and the mask bit is what a future layout player would set.
		r.entries[0].mask |= fieldX
		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none after a layout field", entries)
		}
	})

	t.Run("singular delta", func(t *testing.T) {
		r := New()
		r.now = testClock
		d := scaledDisplay()

		if err := r.Bind(d); err != nil {
			t.Fatalf("bind: %v", err)
		}
		r.Step(d, testClock())
		if entries := r.LayerEntries(); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none at scale(0)", entries)
		}

		r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))
		entries := r.LayerEntries()
		if len(entries) != 1 {
			t.Fatalf("entries = %d, want one once the scale is invertible", len(entries))
		}
		if want := 0.5; !layerFloatClose(entries[0].Matrix.A, want) {
			t.Fatalf("scale delta = %v, want %v", entries[0].Matrix.A, want)
		}
	})

	t.Run("band content", func(t *testing.T) {
		for _, tc := range []struct {
			name string
			set  func(*layout.DisplayOp)
		}{
			{"fixed", func(op *layout.DisplayOp) { op.Fixed = true }},
			{"pinned", func(op *layout.DisplayOp) { op.Pinned = true }},
			{"z-index", func(op *layout.DisplayOp) { op.ZIndex, op.ZIndexSet = 2, true }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				r := New()
				r.now = testClock
				d := animatedDisplay("running", true)

				if err := r.Bind(d); err != nil {
					t.Fatalf("bind: %v", err)
				}
				r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))
				if got := len(r.LayerEntries()); got != 1 {
					t.Fatalf("entries before the band = %d, want one", got)
				}

				tc.set(&d.Ops[0])
				if entries := r.LayerEntries(); len(entries) != 0 {
					t.Fatalf("entries = %+v, want none across a %s band", entries, tc.name)
				}
			})
		}
	})
}

// TestLayerEntriesReadAllocFree pins that a per-frame promotion read reuses the
// runtime's slice and stays off the garbage collector.
func TestLayerEntriesReadAllocFree(t *testing.T) {
	r := New()
	r.now = testClock
	d := layerPairDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))
	if got := len(r.LayerEntries()); got != 2 {
		t.Fatalf("warm entries = %d, want two", got)
	}

	allocs := testing.AllocsPerRun(100, func() {
		_ = r.LayerEntries()
	})
	if allocs != 0 {
		t.Fatalf("a live layer read allocated %.2f objects per run, want 0", allocs)
	}
}

// TestLayerEntriesNilRuntime keeps the accessor safe before a page publishes.
func TestLayerEntriesNilRuntime(t *testing.T) {
	var r *Runtime
	if entries := r.LayerEntries(); entries != nil {
		t.Fatalf("nil runtime entries = %+v, want nil", entries)
	}

	r = New()
	if entries := r.LayerEntries(); entries != nil {
		t.Fatalf("unbound runtime entries = %+v, want nil", entries)
	}
}

// transitionLayerPair builds the old and new display of one chip whose
// transition declarations the new box carries.
func transitionLayerPair(property string, fromOpacity, toOpacity, fromX, toX float64) (
	old, next *layout.Display) {
	records := []layout.Transition{linearTransition(property)}

	return transitionChip(0, fromOpacity, fromX, records...),
		transitionChip(0, toOpacity, toX, records...)
}

// bindTransitionLayer captures and binds one transition between two displays.
func bindTransitionLayer(t *testing.T, r *Runtime, old, next *layout.Display,
	start time.Time) {
	t.Helper()

	captures := r.CaptureTransitions(old, next)
	if err := r.Bind(next); err != nil {
		t.Fatalf("bind animations: %v", err)
	}
	r.BindTransitions(next, captures, start)
}

// layerEntryMatrixBase returns the inverse a page would hand the renderer for
// entry, mirroring page.LayerSpecs.
func layerEntryMatrixBase(t *testing.T, entry LayerEntry) layout.Matrix2D {
	t.Helper()

	base, ok := layerInverse(entry.Matrix)
	if !ok {
		t.Fatalf("entry matrix %+v is not invertible", entry.Matrix)
	}

	return base
}

// TestTransitionLayerTransformPromotes pins the transform transition record: a
// live run becomes one layer entry at the element's box index, with the sampled
// element matrix. The matrix is not a delta from the run's start value, because
// a reversal replaces that value; the inverse bake instead holds the
// element-local layout pose, which is stable for the whole run.
func TestTransitionLayerTransformPromotes(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("transform", 1, 1, 10, 20)

	bindTransitionLayer(t, r, old, next, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))

	entries := r.LayerEntriesFor(next)
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one transform layer", entries)
	}
	entry := entries[0]
	if entry.ElementID != transitionChipID || entry.BoxIndex != 0 {
		t.Fatalf("entry = %q at %d, want %q at 0", entry.ElementID, entry.BoxIndex,
			transitionChipID)
	}
	if !entry.Has2DTransform || entry.HasOpacity {
		t.Fatalf("entry fields = %+v, want a transform entry only", entry)
	}
	if want := layerTranslate(15, 0); !layerMatrixClose(entry.Matrix, want) {
		t.Fatalf("matrix = %+v, want the sampled midpoint %+v", entry.Matrix, want)
	}
	if entry.Alpha != 1 {
		t.Fatalf("alpha = %v, want 1 for a transform-only transition", entry.Alpha)
	}
	if entry.Mode != LayerModeNormal {
		t.Fatalf("mode = %q, want normal", entry.Mode)
	}
	// Matrix composed with the base the page passes is the identity, so the
	// texture holds one fixed pose.
	if got := entry.Matrix.Mul(layerEntryMatrixBase(t, entry)); !layerMatrixClose(got, layout.IdentityMatrix()) {
		t.Fatalf("matrix * base = %+v, want the identity", got)
	}
	if got := next.Ops[0].Transform(); !layerMatrixClose(got, layerTranslate(15, 0)) {
		t.Fatalf("sampled op matrix = %+v, want the midpoint", got)
	}
}

// TestTransitionLayerOpacityPromotes pins group alpha for an opacity
// transition: the entry reads the sampled element opacity, not the product with
// the operation's own source alpha, so the renderer applies it once.
func TestTransitionLayerOpacityPromotes(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("opacity", 1, .5, 0, 0)
	old.Ops[0].Alpha, next.Ops[0].Alpha = .5, .5

	bindTransitionLayer(t, r, old, next, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))

	entries := r.LayerEntriesFor(next)
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one opacity layer", entries)
	}
	entry := entries[0]
	if !entry.HasOpacity || entry.Has2DTransform {
		t.Fatalf("entry fields = %+v, want an opacity entry only", entry)
	}
	if !entry.Matrix.IsIdentity() {
		t.Fatalf("matrix = %+v, want the identity", entry.Matrix)
	}
	if want := 0.75; !layerFloatClose(entry.Alpha, want) {
		t.Fatalf("alpha = %v, want the sampled %v, not the per-op product .375", entry.Alpha, want)
	}
	if got := Read(&next.Ops[0]).Opacity; !near(got, .75) {
		t.Fatalf("sampled op opacity = %v, want .75", got)
	}
}

// TestTransitionLayerReleasedAtEnd checks the release rule at the motion level:
// a finished run leaves the registry in the same Step, so the next promotion
// read names no layer and the renderer's demotion can release the texture.
func TestTransitionLayerReleasedAtEnd(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("transform", 1, 1, 0, 10)

	bindTransitionLayer(t, r, old, next, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))
	if got := len(r.LayerEntriesFor(next)); got != 1 {
		t.Fatalf("mid-run entries = %d, want one", got)
	}

	r.Step(next, start.Add(transitionFrameMS*time.Millisecond))
	if entries := r.LayerEntriesFor(next); len(entries) != 0 {
		t.Fatalf("entries after the end = %+v, want none", entries)
	}
	if got := r.Stats().TransitionsActive; got != 0 {
		t.Fatalf("active transitions after the end = %d, want 0", got)
	}
	if got := len(r.transitions); got != 0 {
		t.Fatalf("registry entries after the end = %d, want 0", got)
	}
}

// TestTransitionLayerReleasedOnCancel checks the other release path: a publish
// that drops the element's transition declaration cancels the run, and the
// promotion read stops naming its layer.
func TestTransitionLayerReleasedOnCancel(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("opacity", 1, .5, 0, 0)

	bindTransitionLayer(t, r, old, next, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))
	if got := len(r.LayerEntriesFor(next)); got != 1 {
		t.Fatalf("mid-run entries = %d, want one", got)
	}

	next.Boxes[0].Transitions = nil
	r.BindTransitions(next, nil, start.Add(transitionFrameMS/2*time.Millisecond))
	if got := r.Stats().TransitionsCancelled; got != 1 {
		t.Fatalf("cancelled = %d, want one", got)
	}
	if entries := r.LayerEntriesFor(next); len(entries) != 0 {
		t.Fatalf("entries after the cancel = %+v, want none", entries)
	}
}

// TestTransitionLayerSharesTheAnimationLayer pins the conflict rule: an
// opacity transition and a transform animation on one element produce one
// entry, one box, one composite matrix, and one group alpha.
func TestTransitionLayerSharesTheAnimationLayer(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("opacity", 1, .5, 0, 0)
	next.Boxes[0].Animations = []layout.Animation{{
		Name: "slide", DurationMS: testDurationMS, IterationInfinite: true,
		FillMode: "both",
	}}
	next.Keyframes = []layout.KeyframesRule{{
		Name: "slide",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: "transform", Value: "translateX(0px)"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: "transform", Value: "translateX(40px)"},
			}},
		},
	}}

	bindTransitionLayer(t, r, old, next, start)
	r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))

	entries := r.LayerEntriesFor(next)
	if len(entries) != 1 {
		t.Fatalf("entries = %+v, want one shared layer", entries)
	}
	entry := entries[0]
	if entry.BoxIndex != 0 {
		t.Fatalf("box = %d, want the single chip box", entry.BoxIndex)
	}
	if !entry.Has2DTransform || !entry.HasOpacity {
		t.Fatalf("entry fields = %+v, want both the animation and transition fields", entry)
	}
	if want := layerTranslate(20, 0); !layerMatrixClose(entry.Matrix, want) {
		t.Fatalf("matrix = %+v, want the animation midpoint %+v", entry.Matrix, want)
	}
	if want := .75; !layerFloatClose(entry.Alpha, want) {
		t.Fatalf("alpha = %v, want the transition midpoint %v", entry.Alpha, want)
	}
}

// TestTransitionLayerProjectionSkips covers the entries a layer cannot carry:
// a live opacity transition on a perspective element is refused with a reason,
// and a transform transition on a 3D element is refused at capture.
func TestTransitionLayerProjectionSkips(t *testing.T) {
	t.Run("perspective opacity", func(t *testing.T) {
		r := New()
		start := testClock()
		r.now = func() time.Time { return start }
		old, next := transitionLayerPair("opacity", 1, .5, 0, 0)
		old.Boxes[0].Transform.HasPerspective = true
		next.Boxes[0].Transform.HasPerspective = true

		bindTransitionLayer(t, r, old, next, start)
		r.Step(next, start.Add(transitionFrameMS/2*time.Millisecond))

		if entries := r.LayerEntriesFor(next); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none for a perspective element", entries)
		}
		skips := r.LayerSkipsFor(next)
		if len(skips) != 1 {
			t.Fatalf("skips = %+v, want the projection reason", skips)
		}
		if skips[0].Reason != layerSkipProjection || skips[0].Property != "opacity" {
			t.Fatalf("skip = %+v, want %q for opacity", skips[0], layerSkipProjection)
		}
	})

	t.Run("3d transform capture", func(t *testing.T) {
		r := New()
		start := testClock()
		r.now = func() time.Time { return start }
		old, next := transitionLayerPair("transform", 1, 1, 0, 10)
		old.Boxes[0].Transform.HasTransform3D = true
		next.Boxes[0].Transform.HasTransform3D = true

		captures := r.CaptureTransitions(old, next)
		if len(captures) != 0 {
			t.Fatalf("3D captures = %d, want none", len(captures))
		}
		r.BindTransitions(next, captures, start)
		if got := r.Stats().TransitionsSkipped; got != 1 {
			t.Fatalf("capture skips = %d, want the 3D reason counted", got)
		}
		if entries := r.LayerEntriesFor(next); len(entries) != 0 {
			t.Fatalf("entries = %+v, want none", entries)
		}
	})
}

// TestLayerEntriesForTransitionsWithAnimationsOff pins the independent
// switches: Bind(nil) releases the animation registry, but a live transition
// still promotes against the display the page hands the read.
func TestLayerEntriesForTransitionsWithAnimationsOff(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }
	old, next := transitionLayerPair("transform", 1, 1, 0, 10)

	captures := r.CaptureTransitions(old, next)
	if err := r.Bind(next); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.BindTransitions(next, captures, start)
	if err := r.Bind(nil); err != nil {
		t.Fatalf("release animations: %v", err)
	}
	r.StepTransitions(next, start.Add(transitionFrameMS/2*time.Millisecond))

	if entries := r.LayerEntries(); len(entries) != 0 {
		t.Fatalf("bound-display entries = %+v, want none after Bind(nil)", entries)
	}
	entries := r.LayerEntriesFor(next)
	if len(entries) != 1 || !entries[0].Has2DTransform {
		t.Fatalf("entries = %+v, want the transition layer", entries)
	}
}

// TestLayerTransitionDisplayCrossFadePending keeps the follow-up honest: an
// element that leaves the display with a display transition is reported as
// blocked with a reason by the transition capture, and it never becomes a layer
// entry, so no cross-fade is faked and no layer is invented for the outgoing
// element.
func TestLayerTransitionDisplayCrossFadePending(t *testing.T) {
	r := New()
	start := testClock()
	r.now = func() time.Time { return start }

	old := transitionChip(0, 1, 0,
		layout.Transition{Property: "display", DurationMS: transitionFrameMS})
	next := &layout.Display{Width: testWidth, Height: testHeight}

	captures := r.CaptureTransitions(old, next)
	if len(captures) != 0 {
		t.Fatalf("display captures = %d, want none", len(captures))
	}
	r.BindTransitions(next, captures, start)

	stats := r.Stats()
	if stats.TransitionsBlocked != 1 {
		t.Fatalf("blocked = %d, want the display exit reported", stats.TransitionsBlocked)
	}
	if !strings.Contains(stats.TransitionSkipReason, "display") {
		t.Fatalf("blocked reason = %q, want a display reason", stats.TransitionSkipReason)
	}
	if entries := r.LayerEntriesFor(old); len(entries) != 0 {
		t.Fatalf("entries on the outgoing display = %+v, want none", entries)
	}
	if entries := r.LayerEntriesFor(next); len(entries) != 0 {
		t.Fatalf("entries on the incoming display = %+v, want none", entries)
	}
}
