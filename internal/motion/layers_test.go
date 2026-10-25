package motion

import (
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
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
