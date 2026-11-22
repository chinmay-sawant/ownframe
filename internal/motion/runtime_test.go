package motion

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"image"
	"strings"
	"testing"
	"time"
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

// TestPromotedLayerStopsStepDamage pins the compositor seam: an element the
// renderer holds as a retained layer writes into its texture, so Step stops
// damaging the page background. Clearing the set (a demotion) restores the
// damage for the next sample.
func TestPromotedLayerStopsStepDamage(t *testing.T) {
	t.Parallel()

	r := New()
	r.now = testClock
	d := animatedDisplay("running", true)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(d, testClock().Add(10*time.Millisecond))

	r.SetPromotedLayers([]string{testChipID})
	damage := r.Step(d, testClock().Add(30*time.Millisecond))
	if !damage.Empty() {
		t.Fatalf("promoted damage = %v, want none", damage.Bounds())
	}

	r.SetPromotedLayers(nil)
	damage = r.Step(d, testClock().Add(50*time.Millisecond))
	if damage.Empty() {
		t.Fatal("demoted entry reported no damage")
	}
}

// splitDisplay is one card with a transform animation on its subtree and a
// background-colour animation on its own fill. The two properties bind as two
// entries that share the fill operation, which is the case the pause handoff
// has to split by paint mask.
func splitDisplay(playState string) *layout.Display {
	return &layout.Display{
		Width: testWidth, Height: testHeight, PointsPerPixel: 1, PixelPerPoint: 1,
		Ops: []layout.DisplayOp{
			{ID: 1, Kind: layout.DisplayOpFillRect,
				X: 10, Y: 10, W: 20, H: 20, R: 0.1, G: 0.2, B: 0.3, Alpha: 1},
			{ID: 2, Kind: layout.DisplayOpText,
				X: 12, Y: 12, W: 10, H: 10, R: 1, G: 1, B: 1, Alpha: 1},
		},
		OpOwners: []int{0, 0},
		Boxes: []layout.Box{{
			ID: "card", Tag: "div", X: 10, Y: 10, W: 20, H: 20, ParentIndex: -1,
			OperationIDs: []uint64{1, 2},
			Animations: []layout.Animation{
				{Name: "slide", DurationMS: testDurationMS, IterationInfinite: true,
					PlayState: playState, FillMode: "both"},
				{Name: "tint", DurationMS: testDurationMS, IterationInfinite: true,
					PlayState: playState, FillMode: "both"},
			},
		}},
		Keyframes: []layout.KeyframesRule{
			{Name: "slide", Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "transform", Value: "translateX(0px)"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "transform", Value: "translateX(20px)"}}},
			}},
			{Name: "tint", Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(0,0,0)"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{
					{Property: "background-color", Value: "rgb(255,255,255)"}}},
			}},
		},
	}
}

// TestPausedRebindKeepsTheDisplayedPose is the page bug in miniature: a redraw
// while paused publishes a display with the declared paint, and the new player
// must not snap the chip back to it before the next tick.
func TestPausedRebindKeepsTheDisplayedPose(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(8); got != want {
		t.Fatalf("sample before pause = %v, want %v", got, want)
	}

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("rebind: %v", err)
	}

	// The redraw happens between ticks, so the pose must already be copied.
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("pose immediately after the paused publish = %v, want %v", got, want)
	}

	r.Step(paused, testClock().Add(500*time.Millisecond))
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused sample = %v, want the frozen %v", got, want)
	}
	if r.Active() || r.Stats().Active != 0 {
		t.Fatalf("a paused runtime reports active=%v count=%d", r.Active(), r.Stats().Active)
	}

	if damage := r.Step(paused, testClock().Add(600*time.Millisecond)); !damage.Empty() {
		t.Fatalf("a second paused tick dirtied %v", damage.Bounds())
	}
}

// TestPausedPoseSurvivesTwoPublishesBeforeATick covers a redraw that follows
// another redraw while the animation is still paused: the second published
// display is built from the declared paint, but the frozen pose must come from
// the first paused display rather than the base.
func TestPausedPoseSurvivesTwoPublishesBeforeATick(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("first pause rebind: %v", err)
	}

	// No Step between the two publishes, so the first paused entry was never
	// sampled; its operations still carry the copied pose.
	again := animatedDisplay("paused", true)
	if err := r.BindAfter(paused, again); err != nil {
		t.Fatalf("second pause rebind: %v", err)
	}
	if got, want := movedBy(&again.Ops[0]), float64(8); got != want {
		t.Fatalf("pose after the second paused publish = %v, want %v", got, want)
	}
}

// TestResumeContinuesFromThePausedPhase resumes 500 ms after the pause and
// compares the sample with an uninterrupted run at the same active time.
func TestResumeContinuesFromThePausedPhase(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))

	paused := animatedDisplay("paused", true)
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("pause rebind: %v", err)
	}
	r.Step(paused, testClock().Add(540*time.Millisecond))

	running := animatedDisplay("running", true)
	r.now = func() time.Time { return testClock().Add(540 * time.Millisecond) }
	if err := r.BindAfter(paused, running); err != nil {
		t.Fatalf("resume rebind: %v", err)
	}
	r.Step(running, testClock().Add(560*time.Millisecond))
	if got, want := movedBy(&running.Ops[0]), float64(12); got != want {
		t.Fatalf("resumed sample = %v, want %v after 60 ms of active time", got, want)
	}

	uninterrupted := New()
	uninterrupted.now = testClock
	fresh := animatedDisplay("running", true)
	if err := uninterrupted.Bind(fresh); err != nil {
		t.Fatalf("bind uninterrupted: %v", err)
	}
	uninterrupted.Step(fresh, testClock().Add(60*time.Millisecond))
	if got, want := movedBy(&fresh.Ops[0]), movedBy(&running.Ops[0]); got != want {
		t.Fatalf("uninterrupted active-time sample = %v, resumed = %v", got, want)
	}
}

// TestPausedPoseCopiesOnlyTheFieldsEachPlayerOwns checks the split handoff: the
// transform entry carries the matrix and the colour entry carries the fill
// channels, and neither overwrites the other or the label text.
func TestPausedPoseCopiesOnlyTheFieldsEachPlayerOwns(t *testing.T) {
	r := New()
	r.now = testClock
	first := splitDisplay("running")

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(40*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(8); got != want {
		t.Fatalf("fill translation = %v, want %v", got, want)
	}
	if got, want := first.Ops[0].R, 0.4; got != want {
		t.Fatalf("fill red = %v, want %v", got, want)
	}

	paused := splitDisplay("paused")
	r.now = func() time.Time { return testClock().Add(80 * time.Millisecond) }
	if err := r.BindAfter(first, paused); err != nil {
		t.Fatalf("rebind: %v", err)
	}

	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused fill translation = %v, want %v", got, want)
	}
	if got, want := paused.Ops[0].R, 0.4; got != want {
		t.Fatalf("paused fill red = %v, want %v", got, want)
	}
	if got, want := movedBy(&paused.Ops[1]), float64(8); got != want {
		t.Fatalf("paused label translation = %v, want %v", got, want)
	}
	if paused.Ops[1].R != 1 || paused.Ops[1].G != 1 || paused.Ops[1].B != 1 {
		t.Fatalf("paused label colour = %v,%v,%v, want its own 1,1,1",
			paused.Ops[1].R, paused.Ops[1].G, paused.Ops[1].B)
	}

	r.Step(paused, testClock().Add(500*time.Millisecond))
	if got, want := movedBy(&paused.Ops[0]), float64(8); got != want {
		t.Fatalf("paused sampled translation = %v, want %v", got, want)
	}
	if got, want := paused.Ops[0].R, 0.4; got != want {
		t.Fatalf("paused sampled red = %v, want %v", got, want)
	}
}

// delayDisplay is animatedDisplay with the cycle and signed start delay the
// pause drift needs.
func delayDisplay(playState string, durationMS, delayMS float64) *layout.Display {
	d := animatedDisplay(playState, true)
	d.Boxes[0].Animations[0].DurationMS = durationMS
	d.Boxes[0].Animations[0].DelayStartMS = delayMS

	return d
}

// sampledDisplay is the display a re-projection patch hands back. The engine
// rebuilds it from the retained layout result, and the previous sample wrote
// through that result's operations, so the list still carries the sampled
// paint instead of the declared one. The copy is what makes the test model
// that handoff: writing through the next display must not move the previous
// one.
func sampledDisplay(d *layout.Display) *layout.Display {
	if d == nil {
		return nil
	}
	clone := *d
	clone.Ops = append([]layout.DisplayOp(nil), d.Ops...)
	clone.OpOwners = append([]int(nil), d.OpOwners...)

	return &clone
}

// TestPausedNegativeDelayHoldsAcrossSampledRebinds is the drift case: a paused
// animation whose negative delay already advanced it a quarter of the cycle is
// rebound five times onto displays that still carry its sample. The frozen
// pose must not move.
func TestPausedNegativeDelayHoldsAcrossSampledRebinds(t *testing.T) {
	r := New()
	r.now = testClock
	first := delayDisplay("paused", 4000, -1000)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(5); got != want {
		t.Fatalf("first paused sample = %v, want %v at a quarter of the shift", got, want)
	}

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))
		if got, want := movedBy(&next.Ops[0]), float64(5); got != want {
			t.Fatalf("rebind %d pose = %v, want the frozen %v", i, got, want)
		}
		previous = next
	}
}

// TestPausedPositiveDelayHoldsAcrossSampledRebinds keeps the other sign
// honest: while the animation is still inside its delay the frozen pose is the
// backwards fill, and a rebind must not advance it into the active interval.
func TestPausedPositiveDelayHoldsAcrossSampledRebinds(t *testing.T) {
	r := New()
	r.now = testClock
	first := delayDisplay("paused", 4000, 1000)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got := movedBy(&first.Ops[0]); got != 0 {
		t.Fatalf("paused delay pose = %v, want the backwards fill 0", got)
	}

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))
		if got := movedBy(&next.Ops[0]); got != 0 {
			t.Fatalf("rebind %d pose = %v, want the delayed 0", i, got)
		}
		previous = next
	}
}

// TestRunningRebindOnSampledDisplayKeepsPhase is the same handoff without a
// pause: a running player bound to a display that still carries its last
// sample must advance by the elapsed time, not by the sample again.
func TestRunningRebindOnSampledDisplayKeepsPhase(t *testing.T) {
	r := New()
	r.now = testClock
	first := animatedDisplay("running", true)

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(20*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(4); got != want {
		t.Fatalf("first sample = %v, want %v", got, want)
	}

	previous := first
	for i, want := range []float64{6, 8, 10, 12} {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(30+10*i) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at)
		if got := movedBy(&next.Ops[0]); got != want {
			t.Fatalf("rebind %d pose = %v, want %v at the same active time", i, got, want)
		}
		previous = next
	}
}

// A hover patch and a layout animation can each publish a display before
// the paint step samples again. Both preserve the last painted transform.
func TestRunningPaintSurvivesTwoRebindsBeforeSample(t *testing.T) {
	r := New()
	r.now = testClock
	display := animatedDisplay("running", true)
	if err := r.Bind(display); err != nil {
		t.Fatal(err)
	}

	start := testClock()
	r.Step(display, start.Add(20*time.Millisecond))
	for range 2 {
		next := sampledDisplay(display)
		if err := r.BindAfter(display, next); err != nil {
			t.Fatal(err)
		}
		display = next
	}

	r.Step(display, start.Add(30*time.Millisecond))
	if got := movedBy(&display.Ops[0]); got != 6 {
		t.Fatalf("translate = %g after two rebinds, want 6", got)
	}
}

// restampDisplay is a page with two paused animations: a 3D rotateY card that
// asks for the engine's re-projection pass every tick, and a 2D chip with a
// negative delay whose pose must not drift when that pass rebinds it.
func restampDisplay() *layout.Display {
	d := delayDisplay("paused", 4000, -1000)
	d.Boxes = append(d.Boxes, layout.Box{
		ID: "flip", Tag: "div", X: 50, Y: 50, W: 20, H: 20, ParentIndex: -1,
		OperationIDs: []uint64{testOpID + 1},
		Animations: []layout.Animation{{
			Name: "flip", DurationMS: 4000, DelayStartMS: -1000,
			IterationInfinite: true, FillMode: "both", PlayState: "paused",
		}},
	})
	d.Ops = append(d.Ops, layout.DisplayOp{
		ID: testOpID + 1, Kind: layout.DisplayOpFillRect,
		X: 50, Y: 50, W: 20, H: 20, R: 1,
	})
	d.OpOwners = append(d.OpOwners, 1)
	d.Keyframes = append(d.Keyframes, layout.KeyframesRule{
		Name: "flip",
		Keyframes: []layout.Keyframe{
			{Offset: 0, Declarations: []layout.KeyframeDeclaration{
				{Property: testTransform, Value: "rotateY(0deg)"}}},
			{Offset: 1, Declarations: []layout.KeyframeDeclaration{
				{Property: testTransform, Value: "rotateY(180deg)"}}},
		},
	})

	return d
}

// TestPaused3DRestampDoesNotDriftThePausedPose rebinds the whole page on every
// tick, as the page's re-projection pass does. The paused 2D chip keeps its
// pose, and the 3D card keeps asking for the same re-projection without writing
// a stale two-dimensional matrix.
func TestPaused3DRestampDoesNotDriftThePausedPose(t *testing.T) {
	r := New()
	r.now = testClock
	first := restampDisplay()

	if err := r.Bind(first); err != nil {
		t.Fatalf("bind: %v", err)
	}
	r.Step(first, testClock().Add(10*time.Millisecond))
	if got, want := movedBy(&first.Ops[0]), float64(5); got != want {
		t.Fatalf("first chip pose = %v, want %v", got, want)
	}
	requests := r.TakeRestamp()
	if len(requests) != 1 || requests[0].ElementID != "flip" || requests[0].Transform == "" {
		t.Fatalf("first restamp requests = %+v, want the flip card's transform", requests)
	}
	want := requests[0].Transform

	previous := first
	for i := 0; i < 5; i++ {
		next := sampledDisplay(previous)
		at := testClock().Add(time.Duration(10+i*10) * time.Millisecond)
		r.now = func() time.Time { return at }
		if err := r.BindAfter(previous, next); err != nil {
			t.Fatalf("rebind %d: %v", i, err)
		}
		r.Step(next, at.Add(time.Millisecond))

		if got := movedBy(&next.Ops[0]); got != 5 {
			t.Fatalf("rebind %d chip pose = %v, want the frozen 5", i, got)
		}
		if next.Ops[1].XformSet {
			t.Fatalf("rebind %d wrote a stale 2D matrix for the 3D card: %+v",
				i, next.Ops[1].Transform())
		}
		requests = r.TakeRestamp()
		if len(requests) != 1 {
			t.Fatalf("rebind %d restamp requests = %+v, want one", i, requests)
		}
		if requests[0].Transform != want {
			t.Fatalf("rebind %d restamp transform = %q, want the frozen %q",
				i, requests[0].Transform, want)
		}
		previous = next
	}
}

// TestRuntimeFinishedPlayerParksItsDamage pins the settled-frame rule: a
// player that did no work holds the pose it already presented, so only its
// first resting sample marks damage. A repeated rest pose marks nothing, and
// a rebind repaints from base once.
func TestRuntimeFinishedPlayerParksItsDamage(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	// A mid-run sample moves and marks.
	mid := r.Step(d, testClock().Add(50*time.Millisecond))
	if mid.Empty() {
		t.Fatal("mid-run step marked no damage")
	}
	if got := r.Stats().Active; got != 1 {
		t.Fatalf("mid-run active = %d, want 1", got)
	}

	// The finishing sample still marks: the end pose is new paint.
	end := r.Step(d, testClock().Add(150*time.Millisecond))
	if end.Empty() {
		t.Fatal("finishing step marked no damage")
	}
	if got := r.Stats().Active; got != 0 {
		t.Fatalf("finished active = %d, want 0", got)
	}

	// Repeated resting samples paint the same pose and mark nothing.
	for i := range 3 {
		rest := r.Step(d, testClock().Add(time.Duration(200+i*50)*time.Millisecond))
		if !rest.Empty() {
			t.Fatalf("rest step %d marked %v, want no damage", i, rest.Bounds())
		}
	}

	// A rebind repaints from the declared base once, then parks again.
	rebound := animatedDisplay("running", false)
	if err := r.Bind(rebound); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	first := r.Step(rebound, testClock().Add(300*time.Millisecond))
	if first.Empty() {
		t.Fatal("first step after rebind marked no damage")
	}
	second := r.Step(rebound, testClock().Add(350*time.Millisecond))
	if !second.Empty() {
		t.Fatalf("second step after rebind marked %v, want no damage", second.Bounds())
	}
}

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
		mask     Mask
		want     fieldMask
		geometry bool
	}{
		{MaskRGB, fieldRGB, false},
		{MaskOpacity, fieldOpacity, false},
		{MaskMatrix, fieldMatrix, true},
		{MaskX | MaskY, fieldX | fieldY, true},
		{MaskAll, maskAll, true},
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

// ResolveTriggersForTest exposes resolveTriggers to package motion_test.
var ResolveTriggersForTest = resolveTriggers
