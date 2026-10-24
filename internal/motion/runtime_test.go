package motion

import (
	"image"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

const (
	testEpoch      = 1000
	testWidth      = 200
	testHeight     = 120
	testChipID     = "chip"
	testOpID       = 7
	testChipX      = 10
	testChipY      = 10
	testChipSize   = 20
	testShift      = 20
	testDurationMS = 100
	testAnimation  = "slide"
	testFade       = "fade"
	testOpacity    = "opacity"
	testShiftValue = "translate(20px)"
	testStartValue = "translate(0px)"
	testTransform  = "transform"
)

func testClock() time.Time { return time.Unix(testEpoch, 0) }

// movedBy is how far the sampled operation matrix moves the chip from its
// layout position. A transform-list keyframe writes the matrix, not the paint
// X, so the driver tests assert the matrix translation.
func movedBy(op *layout.DisplayOp) float64 {
	if !op.XformSet {
		return 0
	}
	return op.Transform().E
}

// animatedDisplay builds the smallest hand-made display the driver needs:
// one fill op owned by one box, with a running or paused translate
// animation from 0 to testShift.
func animatedDisplay(playState string, infinite bool) *layout.Display {
	return &layout.Display{
		Width:          testWidth,
		Height:         testHeight,
		PointsPerPixel: 1,
		PixelPerPoint:  1,
		Ops: []layout.DisplayOp{{
			ID: testOpID, Kind: layout.DisplayOpFillRect,
			X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
			R: 1,
		}},
		OpOwners: []int{0},
		Boxes: []layout.Box{{
			ID: testChipID, Tag: "div",
			X: testChipX, Y: testChipY, W: testChipSize, H: testChipSize,
			OperationIDs: []uint64{testOpID},
			Animations: []layout.Animation{{
				Name: testAnimation, DurationMS: testDurationMS,
				IterationCount: 1, IterationInfinite: infinite,
				PlayState: playState, FillMode: "both",
			}},
		}},
		Keyframes: []layout.KeyframesRule{{
			Name: testAnimation,
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testStartValue},
				}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: testTransform, Value: testShiftValue},
				}},
			},
		}},
	}
}

// opacityDisplay is the allocation-test fixture. It samples a paint field
// directly, so the measurement covers the driver's per-frame work rather
// than the keyframe sampler's transform decomposition.
func opacityDisplay() *layout.Display {
	d := animatedDisplay("running", true)
	d.Boxes[0].Animations[0].Name = testFade
	d.Keyframes = []layout.KeyframesRule{{
		Name: testFade,
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "1"},
			}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: testOpacity, Value: "0.25"},
			}},
		},
	}}

	return d
}

func TestRuntimeBindStepFinish(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	stats := r.Stats()
	if stats.Bound != 1 || stats.LastError != "" {
		t.Fatalf("after bind stats = %+v, want one bound player and no error", stats)
	}

	if !r.Active() {
		t.Fatal("a bound player needs its first sample")
	}

	damage := r.Step(d, testClock().Add(testDurationMS/2*time.Millisecond))
	if got, want := movedBy(&d.Ops[0]), float64(testShift/2); got != want {
		t.Fatalf("mid-step matrix translation = %v, want %v", got, want)
	}

	if got, want := damage.Bounds(), image.Rect(9, 9, 41, 31); got != want {
		t.Fatalf("mid-step damage = %v, want %v", got, want)
	}

	if !r.Active() {
		t.Fatal("a live player reported inactive")
	}

	stats = r.Stats()
	if stats.Active != 1 || stats.Sampled != 1 {
		t.Fatalf("after step stats = %+v, want one active and one sampled", stats)
	}

	damage = r.Step(d, testClock().Add(3*testDurationMS/2*time.Millisecond))
	if got, want := movedBy(&d.Ops[0]), float64(testShift); got != want {
		t.Fatalf("final matrix translation = %v, want %v", got, want)
	}

	if got, want := damage.Bounds(), image.Rect(19, 9, 51, 31); got != want {
		t.Fatalf("finish damage = %v, want %v", got, want)
	}

	if r.Active() {
		t.Fatal("a finished player still reports active")
	}

	stats = r.Stats()
	if stats.Active != 0 || stats.Sampled != 1 {
		t.Fatalf("after finish stats = %+v, want zero active and one sampled", stats)
	}
}

func TestRuntimeRebindKeepsPhase(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", false)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(8); got != want {
		t.Fatalf("first display matrix translation = %v, want %v", got, want)
	}

	// A publish builds a fresh display from the declared styles; the old
	// sampled paint is gone, but the entry keeps its original start.
	second := animatedDisplay("running", false)
	r.Step(second, testClock().Add(60*time.Millisecond))
	if got, want := movedBy(&second.Ops[0]), float64(12); got != want {
		t.Fatalf("generation rebind restarted: matrix translation = %v, want %v", got, want)
	}

	if got := r.Stats().Rebinds; got != 1 {
		t.Fatalf("rebinds = %d, want 1", got)
	}

	third := animatedDisplay("running", false)
	if err := r.Bind(third); err != nil {
		t.Fatalf("explicit bind: %v", err)
	}
	r.Step(third, testClock().Add(80*time.Millisecond))
	if got, want := movedBy(&third.Ops[0]), float64(16); got != want {
		t.Fatalf("explicit rebind restarted: matrix translation = %v, want %v", got, want)
	}

	if got := r.Stats().Rebinds; got != 2 {
		t.Fatalf("rebinds = %d, want 2", got)
	}
}

func TestRuntimeRebindResolvesMovedOps(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", false)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))

	// An insertion before the chip moves its operation from index 0 to 1,
	// so a player that kept a raw pointer would paint the wrong operation.
	moved := animatedDisplay("running", false)
	moved.Ops = append([]layout.DisplayOp{{
		ID: testOpID + 1, Kind: layout.DisplayOpFillRect,
		X: testWidth - 1, Y: testHeight - 1, W: 1, H: 1,
	}}, moved.Ops...)

	r.Step(moved, testClock().Add(60*time.Millisecond))
	if got, want := movedBy(&moved.Ops[1]), float64(12); got != want {
		t.Fatalf("rebound target matrix translation = %v, want %v", got, want)
	}

	if got := moved.Ops[0].X; got != float64(testWidth-1) {
		t.Fatalf("unowned operation moved to X = %v", got)
	}
}

func TestRuntimeStepAllocFree(t *testing.T) {
	r := New()
	r.now = testClock
	d := opacityDisplay()

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	now := testClock().Add(50 * time.Millisecond)
	r.Step(d, now)
	if !r.Active() {
		t.Fatal("the allocation fixture stopped animating")
	}

	allocs := testing.AllocsPerRun(100, func() {
		damage := r.Step(d, now)
		for i := 0; i < damage.N; i++ {
			_ = damage.Rects[i]
		}
	})
	if allocs != 0 {
		t.Fatalf("live step allocated %.2f objects per run, want 0", allocs)
	}
}

func TestRuntimePausedPlayerSamplesOnceAndParks(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("paused", false)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	// A paused animation still has a pose to show, so it asks for the one
	// sample that reaches it and then stops asking.
	if !r.Active() {
		t.Fatal("a freshly bound paused player did not ask for its first sample")
	}

	r.Step(d, testClock().Add(50*time.Millisecond))

	if r.Active() {
		t.Fatal("a sampled paused player reports active")
	}

	if got := r.Stats().Active; got != 0 {
		t.Fatalf("active count = %d, want 0", got)
	}

	if got := movedBy(&d.Ops[0]); got != 0 {
		t.Fatalf("paused player wrote matrix translation = %v, want the identity transform", got)
	}

	// The second step must skip the frozen sample instead of dirtying the
	// page for the same pose again.
	if damage := r.Step(d, testClock().Add(60*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a second paused step produced damage %v", damage.Bounds())
	}

	if got := d.Boxes[0].Animations[0].PlayState; got != "paused" {
		t.Fatalf("bind rewrote PlayState to %q", got)
	}
}

func TestRuntimeBindNilReleases(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", true)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	if !r.Active() {
		t.Fatal("bound runtime is idle")
	}

	if err := r.Bind(nil); err != nil {
		t.Fatalf("release: %v", err)
	}

	if r.Active() {
		t.Fatal("released runtime still active")
	}

	if got := r.Stats().Bound; got != 0 {
		t.Fatalf("bound count after release = %d, want 0", got)
	}

	if damage := r.Step(nil, testClock()); !damage.Empty() {
		t.Fatalf("release step produced damage %v", damage.Bounds())
	}
}

func TestRuntimeWithoutAnimationsStaysIdle(t *testing.T) {
	r := New()
	d := &layout.Display{Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1}

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	if r.Active() {
		t.Fatal("a page with no animations reported active")
	}

	if damage := r.Step(d, testClock()); !damage.Empty() {
		t.Fatalf("a page with no animations produced damage %v", damage.Bounds())
	}

	if got := r.Stats().Bound; got != 0 {
		t.Fatalf("bound count = %d, want 0", got)
	}
}
