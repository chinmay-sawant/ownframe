package motion

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

// threeDDisplay replaces the fixture's translate keyframes with a 3D rotation.
func threeDDisplay() *layout.Display {
	d := animatedDisplay("running", false)
	for i := range d.Keyframes[0].Keyframes {
		value := "rotateY(0deg)"
		if i == 1 {
			value = "rotateY(45deg)"
		}
		d.Keyframes[0].Keyframes[i].Declarations[0].Value = value
	}

	return d
}

func TestRuntimeCountsAThreeDKeyframeWithoutWritingAMatrix(t *testing.T) {
	r := New()
	r.now = testClock
	d := threeDDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	if got := r.Stats().NeedsRestamp; got != 1 {
		t.Fatalf("restamp count = %d, want 1", got)
	}
	if d.Ops[0].XformSet {
		t.Fatalf("a 3D keyframe wrote a 2D matrix %+v", d.Ops[0].Transform())
	}
}

func TestRuntimeCountsAThreeDElement(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)
	parent := layout.Matrix2D{A: 1, D: 1, E: 5}
	d.Boxes[0].Transform = layout.TransformInfo{
		Parent: parent, World: parent, HasTransform3D: true,
	}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	if got := r.Stats().NeedsRestamp; got != 1 {
		t.Fatalf("restamp count = %d, want 1", got)
	}
	if d.Ops[0].XformSet {
		t.Fatalf("a 3D element wrote a 2D matrix %+v", d.Ops[0].Transform())
	}
}

func TestRuntimeTwoDimensionalSampleNeedsNoRestamp(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	if got := r.Stats().NeedsRestamp; got != 0 {
		t.Fatalf("restamp count = %d, want 0", got)
	}
	if !d.Ops[0].XformSet {
		t.Fatal("a 2D transform did not write the operation matrix")
	}
	if requests := r.TakeRestamp(); requests != nil {
		t.Fatalf("a 2D page collected restamp requests %+v", requests)
	}
}

// restampDegrees reads the angle of a serialized rotateY list.
func restampDegrees(t *testing.T, transform string) float64 {
	t.Helper()

	const prefix, suffix = "rotateY(", "deg)"
	if !strings.HasPrefix(transform, prefix) || !strings.HasSuffix(transform, suffix) {
		t.Fatalf("transform = %q, want a rotateY(...deg) list", transform)
	}

	degrees, err := strconv.ParseFloat(transform[len(prefix):len(transform)-len(suffix)], 64)
	if err != nil {
		t.Fatalf("transform %q angle: %v", transform, err)
	}

	return degrees
}

func TestRuntimeCollectsOneRequestPerThreeDimensionalElement(t *testing.T) {
	r := New()
	r.now = testClock
	d := threeDDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one per element", len(requests))
	}
	if requests[0].ElementID != testChipID {
		t.Fatalf("request element = %q, want %q", requests[0].ElementID, testChipID)
	}
	if got := restampDegrees(t, requests[0].Transform); math.Abs(got-22.5) > 1e-9 {
		t.Fatalf("request transform = %q, want rotateY(22.5deg)", requests[0].Transform)
	}
	if got := r.Stats().NeedsRestamp; got != 1 {
		t.Fatalf("restamp count = %d, want 1", got)
	}
}

func TestRuntimeCollectsOneRequestForAMultiOperationElement(t *testing.T) {
	r := New()
	r.now = testClock
	d := threeDDisplay()

	// The element owns a second operation, so the player has two targets. The
	// element is still one restamp.
	second := d.Ops[0]
	second.ID = testOpID + 1
	d.Ops = append(d.Ops, second)
	d.OpOwners = append(d.OpOwners, 0)
	d.Boxes[0].OperationIDs = []uint64{testOpID, testOpID + 1}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(50*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one for the element's two operations", len(requests))
	}
}

func TestRuntimeTakeRestampClearsTheList(t *testing.T) {
	r := New()
	r.now = testClock
	d := threeDDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(25*time.Millisecond))
	if len(r.TakeRestamp()) != 1 {
		t.Fatal("the first step collected no request")
	}
	if requests := r.TakeRestamp(); requests != nil {
		t.Fatalf("a second take returned %+v, want the list cleared", requests)
	}

	r.Step(d, testClock().Add(75*time.Millisecond))
	if requests := r.TakeRestamp(); len(requests) != 1 {
		t.Fatalf("the next step collected %d requests, want 1", len(requests))
	}
}

func TestRuntimeBindReportsMotionThatNeedsRetainedState(t *testing.T) {
	r := New()
	r.now = testClock

	if err := r.Bind(threeDDisplay()); err != nil {
		t.Fatalf("bind 3D: %v", err)
	}
	if !r.NeedsRetainedState() {
		t.Fatal("a 3D keyframe did not ask for retained layout state")
	}
	if !r.Stats().NeedsRetainedState {
		t.Fatal("the stats snapshot did not report the 3D bind")
	}

	if err := r.Bind(animatedDisplay("running", false)); err != nil {
		t.Fatalf("bind 2D: %v", err)
	}
	if r.NeedsRetainedState() {
		t.Fatal("a 2D translate page asked for retained layout state")
	}
}

func TestRuntimeNoteRestampRecordsTheOutcome(t *testing.T) {
	r := New()

	r.NoteRestamp(2, nil)
	if stats := r.Stats(); stats.Restamped != 2 || stats.RestampError != "" {
		t.Fatalf("after success stats = %+v, want two restamped and no error", stats)
	}

	r.NoteRestamp(0, errRestampTest)
	if stats := r.Stats(); stats.Restamped != 0 || !strings.Contains(stats.RestampError, errRestampTest.Error()) {
		t.Fatalf("after failure stats = %+v, want the error recorded", stats)
	}
}

// errRestampTest stands in for the engine's invalid-override error.
var errRestampTest = errors.New("layout: invalid transform override")
