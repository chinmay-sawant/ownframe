package motion

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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
	if got := cssmotion.Read(&next.Ops[0]).Opacity; !near(got, .75) {
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
