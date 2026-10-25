package motion

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/csstrigger"
)

// splitTriggerDisplay is one element whose transform and background colour
// bind as two players. A trigger has to reach the colour player even though
// that player holds a single effect while the box lists two animations.
func splitTriggerDisplay() *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{
				ID: testOpID, Kind: layout.DisplayOpFillRect,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize, R: 0, Alpha: 1,
			},
			{
				ID: testOpID + 1, Kind: layout.DisplayOpText,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize, R: 1, Alpha: 1,
			},
		},
		OpOwners: []int{0, 1},
		Boxes: []layout.Box{
			{
				ID: testChipID, Tag: "div", ParentIndex: -1,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
				OperationIDs: []uint64{testOpID, testOpID + 1},
				Animations: []layout.Animation{
					{
						Name: testAnimation, DurationMS: testDurationMS,
						TimingFunction: "linear", IterationCount: 1, FillMode: "both",
						Trigger: "--move play",
					},
					{
						Name: testFade, DurationMS: testDurationMS,
						TimingFunction: "linear", IterationCount: 1, FillMode: "both",
						Trigger: "--tint play",
					},
				},
			},
			{
				ID: "label", Tag: "span", ParentIndex: 0,
				X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
				OperationIDs: []uint64{testOpID + 1},
			},
		},
		Keyframes: []layout.KeyframesRule{
			{Name: testAnimation, Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testStartValue},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testShiftValue},
				}},
			}},
			{Name: testFade, Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(0,0,0)"},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(255,255,255)"},
				}},
			}},
		},
	}
}

// TestTriggerReachesOnePlayerOfASplitElement proves the pass builder aligns a
// box animation index to the effect index of the player it actually built.
// The colour player holds one effect while the box lists two animations; a raw
// index would ask it for effect 1 and fail.
func TestTriggerReachesOnePlayerOfASplitElement(t *testing.T) {
	d := splitTriggerDisplay()
	r := New()
	r.now = testClock
	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if got := r.Stats().Bound; got != 2 {
		t.Fatalf("bound players = %d, want two target sets", got)
	}

	passes := r.TriggerPasses()
	if len(passes) != 2 {
		t.Fatalf("trigger passes = %d, want 2", len(passes))
	}
	now := testClock()
	for i := range passes {
		event := csstrigger.Event{Name: "--tint", Source: 0, Active: true}
		if err := csstrigger.Dispatch(passes[i].Boxes, event, passes[i].Players, now); err != nil {
			t.Fatalf("pass %d: %v", i, err)
		}
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("a colour trigger moved the transform by %v", got)
	}
	if got := d.Ops[0].R; got < 0.49 || got > 0.51 {
		t.Fatalf("colour sample R = %v, want 0.5", got)
	}
}
