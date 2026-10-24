package page

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

const motionSlideFrames = `
@keyframes slide {
	0% { transform: translateX(0px); }
	100% { transform: translateX(120px); }
}`

const (
	motionRunning = `animation-name: slide; animation-duration: 2s;
		animation-timing-function: linear; animation-iteration-count: infinite;
		animation-fill-mode: both;`
	motionPaused = motionRunning + ` animation-play-state: paused;`
)

// newMotionSource builds a page whose one target div declares the given
// animation declarations, without drawing it yet.
func newMotionSource(t *testing.T, animation string) *Page {
	t.Helper()

	source := `<style>` + motionSlideFrames + `
#target { width: 40px; height: 20px; background-color: rgb(20,40,60); ` + animation + ` }
</style><div id="target"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	return p
}

// newMotionPage draws the fixture and drains the first-frame damage.
func newMotionPage(t *testing.T, animation string) *Page {
	t.Helper()

	p := newMotionSource(t, animation)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// motionTargetIndex finds the first display operation the target box owns.
func motionTargetIndex(t *testing.T, p *Page) int {
	t.Helper()

	d := p.Display()
	if d == nil {
		t.Fatal("the fixture painted no display list")
	}

	var target Box
	for _, box := range p.Boxes() {
		if box.ID == "target" {
			target = box

			break
		}
	}

	indices := cssmotion.OperationIndices(d, target)
	if len(indices) == 0 {
		t.Fatal("the fixture target owns no operation")
	}

	return indices[0]
}

// motionTargetX reports the target operation's x after its baked translation.
// A transform-list keyframe writes translate through SetXform, so the op's X
// stays put and the movement lives in the matrix. Every fixture here animates
// a pure translate, so adding E to X is the painted position.
func motionTargetX(p *Page, index int) float64 {
	op := &p.Display().Ops[index]
	return op.X + op.Transform().E
}

func TestTickDrivesCSSAnimation(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	if p.Ticking() {
		t.Fatal("the fixture registered a tick callback")
	}

	index := motionTargetIndex(t, p)
	base := motionTargetX(p, index)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.FrameDirty() {
		t.Fatal("the driver's damage did not set FrameDirty")
	}
	if got := p.Motion().Bound; got != 1 {
		t.Fatalf("bound players = %d, want 1", got)
	}

	first := motionTargetX(p, index)
	time.Sleep(5 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if second := motionTargetX(p, index); second <= first {
		t.Fatalf("tick X = %v, want it past the first tick's %v (base %v)", second, first, base)
	}
}

func TestPausedCSSAnimationDoesNotRequestFrame(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionPaused)
	index := motionTargetIndex(t, p)
	base := motionTargetX(p, index)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !p.NextFrameAt().IsZero() {
		t.Fatal("a paused animation asked for another frame")
	}
	if got := p.Motion().Active; got != 0 {
		t.Fatalf("active players = %d, want 0", got)
	}
	if got := motionTargetX(p, index); got != base {
		t.Fatalf("paused animation wrote X = %v, want the base %v", got, base)
	}
}

func TestRedrawRebindsCSSAnimationWithPhase(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	index := motionTargetIndex(t, p)
	ctx := context.Background()

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	before := motionTargetX(p, index)
	if before <= 0 {
		t.Fatalf("the animation did not advance before the redraw: X = %v", before)
	}

	gen := p.Generation()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Generation() == gen {
		t.Fatal("Redraw did not bump the generation")
	}

	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if after := motionTargetX(p, index); after < before {
		t.Fatalf("rebind restarted the animation: X = %v, want at least the pre-redraw %v", after, before)
	}
	if got := p.Motion().Rebinds; got == 0 {
		t.Fatal("the redraw did not rebind the players")
	}
}

func TestSetAnimationsOffLeavesBasePaint(t *testing.T) {
	t.Parallel()

	p := newMotionSource(t, motionRunning)
	p.SetAnimations(false)
	if p.Animations() {
		t.Fatal("SetAnimations(false) left the driver on")
	}

	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
	index := motionTargetIndex(t, p)
	base := motionTargetX(p, index)

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := motionTargetX(p, index); got != base {
		t.Fatalf("disabled driver wrote X = %v, want the base %v", got, base)
	}
	if p.FrameDirty() {
		t.Fatal("disabled driver left FrameDirty set")
	}
	if !p.NextFrameAt().IsZero() {
		t.Fatal("disabled driver asked for another frame")
	}
	if got := p.Motion().Bound; got != 0 {
		t.Fatalf("bound players = %d, want 0", got)
	}
}

func TestTickWithoutAnimationsStaysClean(t *testing.T) {
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

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if p.FrameDirty() {
		t.Fatal("a page with no animations set FrameDirty")
	}
	if !p.NextFrameAt().IsZero() {
		t.Fatal("a page with no animations asked for another frame")
	}
}
