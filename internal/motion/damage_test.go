package motion

import (
	"errors"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/dirty"
	"image"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDamageCoalescesNeighbours(t *testing.T) {
	var d Damage

	d.Add(image.Rect(0, 0, 10, 10))
	d.Add(image.Rect(12, 0, 22, 10))

	if got := d.N; got != 1 {
		t.Fatalf("rect count = %d, want 1 after coalescing", got)
	}

	if got, want := d.Bounds(), image.Rect(0, 0, 22, 10); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

func TestDamageOverflowCollapsesToBounds(t *testing.T) {
	var d Damage

	for i := range dirty.MaxRects + 1 {
		x := i * 100
		d.Add(image.Rect(x, 0, x+10, 10))
	}

	if !d.Overflow {
		t.Fatal("damage past the cap did not report overflow")
	}

	if got := d.N; got != 1 {
		t.Fatalf("rect count = %d, want the collapsed bounds", got)
	}

	if got, want := d.Bounds(), image.Rect(0, 0, dirty.MaxRects*100+10, 10); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}
}

func TestDamageIgnoresEmptyRects(t *testing.T) {
	var d Damage

	d.Add(image.Rectangle{})
	if !d.Empty() {
		t.Fatalf("empty rect produced damage %v", d.Bounds())
	}
}

// TestStampTransformClipsRecordsAncestorBox pins the overflow-clip handoff:
// an operation a transform player writes carries its owner box's ancestor
// clip, so the replay can intersect transformed paint with it. Entries that
// write no matrix leave their ops alone.
func TestStampTransformClipsRecordsAncestorBox(t *testing.T) {
	d := &layout.Display{
		Width: 200, Height: 100, PointsPerPixel: 0.75, PixelPerPoint: 4.0 / 3.0,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect, X: 10, Y: 10, W: 20, H: 20},
			{ID: 2, Kind: layout.DisplayOpFillRect, X: 50, Y: 50, W: 10, H: 10},
		},
		OpOwners: []int{0, 1},
		Boxes: []layout.Box{
			{ID: "clip", Tag: "div", ParentIndex: -1,
				OverflowClip:    [4]float64{8, 8, 160, 80},
				HasOverflowClip: true,
				OperationIDs:    []uint64{1}},
			{ID: "free", Tag: "div", ParentIndex: -1, OperationIDs: []uint64{2}},
		},
	}

	entries := []entry{
		{mask: fieldMatrix, indices: []int{0}, boxIndex: 0},
		{mask: fieldOpacity, indices: []int{1}, boxIndex: 1},
	}

	stampTransformClips(d, entries)

	if x, y, w, h, ok := d.Ops[0].OverflowClip(); !ok {
		t.Fatal("matrix-driven op has no overflow clip")
	} else if x != 8*0.75 || y != 8*0.75 || w != 160*0.75 || h != 80*0.75 {
		t.Fatalf("clip = %.2f,%.2f %.2fx%.2f, want 6,6 120x60", x, y, w, h)
	}

	if _, _, _, _, ok := d.Ops[1].OverflowClip(); ok {
		t.Fatal("opacity-only op gained an overflow clip")
	}
}

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

// perspectiveDisplay replaces the fixture's translate keyframes with a
// perspective distance animation. PointsPerPixel is 1 in the fixture, so a
// 500px declaration is 500 points.
func perspectiveDisplay() *layout.Display {
	d := animatedDisplay("running", false)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective", Value: "500px",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective", Value: "1000px",
	}

	return d
}

// addOrigin declares the same perspective-origin on both keyframes.
func addOrigin(d *layout.Display, value string) {
	for i := range d.Keyframes[0].Keyframes {
		d.Keyframes[0].Keyframes[i].Declarations = append(d.Keyframes[0].Keyframes[i].Declarations,
			layout.KeyframeDeclaration{Property: "perspective-origin", Value: value})
	}
}

func TestRuntimeCollectsAPerspectiveRequest(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one per element", len(requests))
	}
	request := requests[0]
	if request.ElementID != testChipID {
		t.Fatalf("request element = %q, want %q", request.ElementID, testChipID)
	}
	if !request.HasPerspective || math.Abs(request.Perspective-750) > 1e-9 {
		t.Fatalf("perspective = %+v, want the 750-point midpoint", request)
	}
	if request.Transform != "" {
		t.Fatalf("perspective-only request carried transform %q", request.Transform)
	}
	if request.HasPerspectiveOrigin {
		t.Fatalf("request reported an origin the keyframes did not declare: %+v", request)
	}
	// The keyframe parser resolves the 50%/50% default against the box, and the
	// driver lifts the box-relative offset into the page point the engine wants.
	if math.Abs(request.PerspectiveOriginX-testChipX-testChipSize/2) > 1e-9 ||
		math.Abs(request.PerspectiveOriginY-testChipY-testChipSize/2) > 1e-9 {
		t.Fatalf("default origin = %v,%v, want the box centre in page points",
			request.PerspectiveOriginX, request.PerspectiveOriginY)
	}
	if got := r.Stats().NeedsRestamp; got != 1 {
		t.Fatalf("restamp count = %d, want 1", got)
	}
}

func TestRuntimeCarriesAnOriginOnlyRequest(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)
	d.Keyframes[0].Keyframes[0].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective-origin", Value: "0 0",
	}
	d.Keyframes[0].Keyframes[1].Declarations[0] = layout.KeyframeDeclaration{
		Property: "perspective-origin", Value: "100% 100%",
	}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if request.HasPerspective {
		t.Fatalf("an origin-only sample set HasPerspective: %+v", request)
	}
	if !request.HasPerspectiveOrigin {
		t.Fatalf("an origin-only sample lost the origin: %+v", request)
	}
	if request.PerspectiveOriginX != testChipX+testChipSize/2 ||
		request.PerspectiveOriginY != testChipY+testChipSize/2 {
		t.Fatalf("origin = %v,%v, want the page midpoint of the box",
			request.PerspectiveOriginX, request.PerspectiveOriginY)
	}
}

func TestRuntimeCarriesThePerspectiveOriginAsPageCoordinates(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()
	addOrigin(d, "top left")

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want 1", len(requests))
	}
	request := requests[0]
	if !request.HasPerspectiveOrigin {
		t.Fatalf("request lost the declared origin: %+v", request)
	}
	if request.PerspectiveOriginX != testChipX || request.PerspectiveOriginY != testChipY {
		t.Fatalf("origin = %v,%v, want the box top left %v,%v",
			request.PerspectiveOriginX, request.PerspectiveOriginY, testChipX, testChipY)
	}
}

func TestRuntimeSharesOneRequestForPerspectiveAndThreeD(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()
	for i := range d.Keyframes[0].Keyframes {
		value := "rotateY(0deg)"
		if i == 1 {
			value = "rotateY(90deg)"
		}
		d.Keyframes[0].Keyframes[i].Declarations = append(d.Keyframes[0].Keyframes[i].Declarations,
			layout.KeyframeDeclaration{Property: "transform", Value: value})
	}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	requests := r.TakeRestamp()
	if len(requests) != 1 {
		t.Fatalf("restamp requests = %d, want one request for the element", len(requests))
	}
	if requests[0].Transform != "rotateY(45deg)" {
		t.Fatalf("transform = %q, want the sampled rotateY", requests[0].Transform)
	}
	if !requests[0].HasPerspective {
		t.Fatalf("the shared request lost the perspective: %+v", requests[0])
	}
}

func TestRuntimePerspectiveBindNeedsRetainedState(t *testing.T) {
	r := New()
	r.now = testClock

	if err := r.Bind(perspectiveDisplay()); err != nil {
		t.Fatalf("bind: %v", err)
	}
	if !r.NeedsRetainedState() {
		t.Fatal("a perspective keyframe did not ask for retained layout state")
	}

	r = New()
	r.now = testClock
	if err := r.Bind(animatedDisplay("running", false)); err != nil {
		t.Fatalf("bind 2D: %v", err)
	}
	if r.NeedsRetainedState() {
		t.Fatal("a 2D translate page asked for retained layout state")
	}
}

func TestRuntimePerspectiveOnlyPageCollectsNoMatrixWrite(t *testing.T) {
	r := New()
	r.now = testClock
	d := perspectiveDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))

	if d.Ops[0].XformSet {
		t.Fatal("a perspective-only sample wrote a 2D operation matrix")
	}
}
