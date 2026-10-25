package motion

import (
	"image"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestOpacityAnimationKeepsItsMask pins the field mask of one paint
// property: an opacity animation writes the subtree, and a transform animation
// writes only the matrix, never the colour or the size.
func TestOpacityAnimationKeepsItsMask(t *testing.T) {
	d := animatedDisplay("running", true)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "opacity", Value: "1",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "opacity", Value: "0.25",
	}

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	entry := r.entries[0]
	if entry.mask != fieldOpacity {
		t.Fatalf("opacity mask = %b, want the opacity field only", entry.mask)
	}

	r.Step(d, testClock().Add(50*time.Millisecond))
	if d.Ops[0].XformSet {
		t.Fatal("an opacity animation wrote a transform matrix")
	}
	if got := d.Ops[0].Opacity(); got != .625 {
		t.Fatalf("opacity = %v, want the midpoint .625", got)
	}
}

// TestTranslucentOperationKeepsAlphaOnTransform pins the phase 03 regression: a
// transform animation leaves the operation's alpha, size, and colour alone.
func TestTranslucentOperationKeepsAlphaOnTransform(t *testing.T) {
	const wantAlpha, wantRed, wantW = .4, .2, testChipSize

	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)
	d.Ops[0].Alpha = wantAlpha
	d.Ops[0].R = wantRed

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	if d.Ops[0].Alpha != wantAlpha {
		t.Fatalf("alpha = %v, want the untouched %v", d.Ops[0].Alpha, wantAlpha)
	}
	if d.Ops[0].R != wantRed {
		t.Fatalf("fill red = %v, want the untouched %v", d.Ops[0].R, wantRed)
	}
	if d.Ops[0].W != wantW {
		t.Fatalf("width = %v, want the laid-out %v", d.Ops[0].W, wantW)
	}
	if got := movedBy(&d.Ops[0]); got == 0 {
		t.Fatal("the transform did not move the translucent operation")
	}
}

// opacityDisplayWith builds one opacity animation over a hand-made subtree.
func opacityDisplayWith(boxes []layout.Box, ops []layout.DisplayOp, owners []int) *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: ops, OpOwners: owners, Boxes: boxes,
		Keyframes: []layout.KeyframesRule{{
			Name: testFade,
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: testOpacity, Value: "1"},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: testOpacity, Value: "0.2"},
				}},
			},
		}},
	}
}

func fadeRecord() layout.Animation {
	return layout.Animation{Name: testFade, DurationMS: testDurationMS,
		IterationCount: 1, IterationInfinite: true, PlayState: "running", FillMode: "both"}
}

// TestColourOnlyEntryDamagesItsPaint pins the damage side of a narrowed mask:
// a colour change repaints the operation's painted box, not the element's whole
// box and not the page.
func TestColourOnlyEntryDamagesItsPaint(t *testing.T) {
	d := animatedDisplay("running", false)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "background-color", Value: "#ff0000",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "background-color", Value: "#00ff00",
	}

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if got := r.entries[0].mask; got != fieldRGB {
		t.Fatalf("colour mask = %b, want the colour field only", got)
	}

	damage := r.Step(d, testClock().Add(50*time.Millisecond))
	if got, want := damage.Bounds(), image.Rect(testChipX-1, testChipY-1, testChipX+testChipSize+1, testChipY+testChipSize+1); got != want {
		t.Fatalf("colour damage = %v, want the painted box %v", got, want)
	}
}

// TestOverlappingOpacityIsReported pins the phase 08 hand-off: per-operation
// opacity cannot flatten two overlapping branches of one subtree, so the bind
// reports a fidelity gap instead of claiming parity.
func TestOverlappingOpacityIsReported(t *testing.T) {
	d := opacityDisplayWith(
		[]layout.Box{
			{ID: "stack", X: 10, Y: 10, W: 30, H: 30, OperationIDs: []uint64{1, 2},
				Animations: []layout.Animation{fadeRecord()}, ParentIndex: -1},
			{ID: "a", X: 10, Y: 10, W: 20, H: 20, OperationIDs: []uint64{1}, ParentIndex: 0},
			{ID: "b", X: 15, Y: 15, W: 20, H: 20, OperationIDs: []uint64{2}, ParentIndex: 0},
		},
		[]layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20, R: 1, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpFillRect, X: 15, Y: 15, W: 20, H: 20, R: .5, Alpha: 1},
		},
		[]int{1, 2})

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	gaps := r.Stats().Gaps
	if len(gaps) != 1 || !strings.Contains(gaps[0], "phase 08") {
		t.Fatalf("gaps = %v, want one overlapping-opacity gap", gaps)
	}

	if err := r.Bind(nil); err != nil {
		t.Fatalf("release: %v", err)
	}
	if got := r.Stats().Gaps; len(got) != 0 {
		t.Fatalf("gaps after release = %v, want none", got)
	}
}

// TestContainedOpacityNeedsNoGap checks the other side of the rule: a label on
// its own card's fill stacks in one ancestor chain, so per-operation alpha
// matches a group alpha and no gap is claimed.
func TestContainedOpacityNeedsNoGap(t *testing.T) {
	d := opacityDisplayWith(
		[]layout.Box{{ID: "card", X: 10, Y: 10, W: 30, H: 30,
			OperationIDs: []uint64{1, 2}, Animations: []layout.Animation{fadeRecord()}, ParentIndex: -1}},
		[]layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 30, H: 30, R: 1, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpText, X: 15, Y: 20, W: 10, H: 8, R: 1, Alpha: 1},
		},
		[]int{0, 0})

	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if got := r.Stats().Gaps; len(got) != 0 {
		t.Fatalf("gaps = %v, want none for a label on its own card", got)
	}
}

// TestFieldMaskOfNarrowsGeometry pins the mapping between the paint mask and
// the damage mask.
func TestFieldMaskOfNarrowsGeometry(t *testing.T) {
	tests := []struct {
		mask     cssmotion.Mask
		want     fieldMask
		geometry bool
	}{
		{cssmotion.MaskRGB, fieldRGB, false},
		{cssmotion.MaskOpacity, fieldOpacity, false},
		{cssmotion.MaskMatrix, fieldMatrix, true},
		{cssmotion.MaskX | cssmotion.MaskY, fieldX | fieldY, true},
		{cssmotion.MaskAll, maskAll, true},
		{0, maskAll, true},
	}
	for _, test := range tests {
		got := fieldMaskOf(test.mask)
		if got != test.want {
			t.Errorf("fieldMaskOf(%b) = %b, want %b", test.mask, got, test.want)
		}
		if got.geometry() != test.geometry {
			t.Errorf("fieldMaskOf(%b).geometry() = %v, want %v", test.mask, got.geometry(), test.geometry)
		}
	}
}
