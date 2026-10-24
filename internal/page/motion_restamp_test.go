package page

import (
	"context"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

const restampFlipFrames = `
@keyframes flip {
	0% { transform: rotateY(0deg); }
	100% { transform: rotateY(180deg); }
}`

const restampStyle = `animation-name: flip; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both;`

// newRestampPage draws a page whose one target flips around the Y axis. The
// animation is 3D, so the display has to be restamped to show it.
func newRestampPage(t *testing.T, retain bool) *Page {
	t.Helper()

	source := `<style>` + restampFlipFrames + `
#target { width: 80px; height: 40px; background-color: rgb(20,40,60); ` + restampStyle + ` }
</style><div id="target"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: retain})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// restampMatrix returns the matrix the target's first operation carries.
func restampMatrix(t *testing.T, p *Page) layout.Matrix2D {
	t.Helper()

	op := &p.Display().Ops[motionTargetIndex(t, p)]
	if !op.XformSet {
		return layout.IdentityMatrix()
	}

	return op.Transform()
}

func TestRotateYAnimationRestampsTheProjectionOnEveryTick(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	base := restampMatrix(t, p)
	if !p.retainState {
		t.Fatal("the fixture started without retained layout state")
	}
	if p.restampRedraw {
		t.Fatal("a retained page owed a redraw before its first tick")
	}

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := restampMatrix(t, p)
	if first == base {
		t.Fatalf("the first tick left the identity matrix %+v alone", first)
	}
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("restamped elements = %d, want 1", got)
	}
	if got := p.Motion().RestampError; got != "" {
		t.Fatalf("restamp error = %q, want none", got)
	}

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	second := restampMatrix(t, p)
	if second == first {
		t.Fatalf("the second tick kept the matrix %+v", second)
	}
	// rotateY projects to cos(angle) on the x axis, so the scale has to fall
	// while the flip runs from 0 to 90 degrees.
	if second.A >= first.A {
		t.Fatalf("rotateY scale went from %v to %v, want a smaller value", first.A, second.A)
	}
}

func TestFirstTickEnablesRetainedStateAndSecondTickMoves(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, false)
	if !p.Motion().NeedsRetainedState {
		t.Fatal("the bind did not report the 3D keyframe")
	}
	if !p.retainState {
		t.Fatal("the bind did not turn retained layout state on")
	}
	base := restampMatrix(t, p)
	gen := p.Generation()

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got != 0 {
		t.Fatalf("the first tick restamped %d elements, want the owed redraw instead", got)
	}
	if got := restampMatrix(t, p); got != base {
		t.Fatalf("the first tick moved the projection to %+v, want the first frame still", got)
	}
	if p.Generation() == gen {
		t.Fatal("the first tick did not redraw with retained layout state")
	}
	if p.restampRedraw {
		t.Fatal("the owed redraw stayed queued after the first tick")
	}

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("the second tick restamped %d elements, want 1", got)
	}
	if got := restampMatrix(t, p); got == base {
		t.Fatalf("the second tick kept the base matrix %+v", got)
	}
}

func TestTranslateAnimationStaysOffTheRestampPath(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	if p.Motion().NeedsRetainedState {
		t.Fatal("a 2D translate page asked for retained layout state")
	}
	if p.retainState {
		t.Fatal("a 2D translate page turned retained layout state on")
	}
	gen := p.Generation()

	for i := 0; i < 2; i++ {
		if err := p.Tick(context.Background()); err != nil {
			t.Fatal(err)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if p.Generation() != gen {
		t.Fatal("a 2D page rebuilt its display, so it reached the restamp path")
	}
	if got := p.Motion().Restamped; got != 0 {
		t.Fatalf("a 2D page restamped %d elements", got)
	}
	if got := p.Motion().NeedsRestamp; got != 0 {
		t.Fatalf("a 2D page asked for %d restamps", got)
	}
	if p.retainState {
		t.Fatal("a 2D page turned retained layout state on")
	}
}

func TestRestampAppliesEveryThreeDimensionalElementInOnePass(t *testing.T) {
	t.Parallel()

	source := `<style>` + restampFlipFrames + `
.card { width: 60px; height: 30px; background-color: rgb(20,40,60); ` + restampStyle + ` }
</style><div class="card" id="left"></div><div class="card" id="right"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100, RetainState: true})
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
	if got := p.Motion().Restamped; got != 2 {
		t.Fatalf("restamped elements = %d, want both cards in one pass", got)
	}

	for _, id := range []string{"left", "right"} {
		indices := cssmotion.OperationIndices(p.Display(), p.boxByID(id))
		if len(indices) == 0 {
			t.Fatalf("%s owns no operation", id)
		}
		if got := p.Display().Ops[indices[0]].Transform(); got == (layout.Matrix2D{}) {
			t.Fatalf("%s kept the zero matrix", id)
		}
	}
}

func TestAdoptRestampMarksTheDamageOnce(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	box := image.Rect(10, 10, 90, 50)

	clone := *p.Display()
	p.adoptRestamp(&layout.ElementPatch{Display: &clone, Old: box, New: box})

	if got := len(p.dirty.Rects); got != 1 {
		t.Fatalf("damage rects = %d (%v), want the shared old and new box marked once",
			got, p.dirty.Rects)
	}
	if got := p.dirty.Rects[0]; got != box {
		t.Fatalf("damage rect = %v, want %v", got, box)
	}

	// A projection that changes what the subtree paints marks the space it
	// left and the space it now covers.
	elsewhere := *p.Display()
	p.adoptRestamp(&layout.ElementPatch{Display: &elsewhere,
		Old: image.Rect(10, 10, 90, 50), New: image.Rect(110, 10, 190, 50)})
	if got := len(p.dirty.Rects); got != 2 {
		t.Fatalf("damage rects = %d (%v), want the old and new box", got, p.dirty.Rects)
	}
}

func TestAdoptRestampBumpsTheGenerationAndRebinds(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	gen := p.Generation()

	clone := *p.Display()
	p.adoptRestamp(&layout.ElementPatch{Display: &clone,
		Old: image.Rect(10, 10, 90, 50), New: image.Rect(10, 10, 90, 50)})

	if got := p.Generation(); got != gen+1 {
		t.Fatalf("generation = %d, want %d", got, gen+1)
	}
	if p.Display() != &clone {
		t.Fatal("the page did not adopt the restamped display")
	}
	if got := p.Motion().Bound; got != 1 {
		t.Fatalf("bound players after the restamp = %d, want 1", got)
	}
}

func TestFailingRestampKeepsThePreviousDisplay(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	before := p.Display()
	gen := p.Generation()

	// The engine rejects the unterminated list, so the page must keep drawing
	// the display it already had.
	p.applyRestamp(context.Background(), []motion.Request{{
		ElementID: "target",
		Transform: "rotateY(45deg",
	}})

	if p.Display() != before {
		t.Fatal("a failed override replaced the display")
	}
	if got := p.Generation(); got != gen {
		t.Fatalf("a failed override bumped the generation to %d, want %d", got, gen)
	}
	stats := p.Motion()
	if !strings.Contains(stats.RestampError, "invalid transform override") {
		t.Fatalf("restamp error = %q, want the invalid override recorded", stats.RestampError)
	}
	if got := stats.Restamped; got != 0 {
		t.Fatalf("restamped elements after the failure = %d, want 0", got)
	}
}

func TestRestampNamesAMissingElementWithoutChangingThePage(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	before := p.Display()

	p.applyRestamp(context.Background(), []motion.Request{{
		ElementID: "missing",
		Transform: "rotateY(45deg)",
	}})

	if p.Display() != before {
		t.Fatal("a missing element replaced the display")
	}
	if got := p.Motion().RestampError; !strings.Contains(got, "no element") {
		t.Fatalf("restamp error = %q, want the missing element recorded", got)
	}
}
