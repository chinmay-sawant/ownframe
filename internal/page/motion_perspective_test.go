package page

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// perspectiveCameraFrames animates the stage's camera distance and leaves the
// card's own rotateY static, so only the projection changes.
const perspectiveCameraFrames = `
@keyframes camera {
	0% { perspective: 500px; }
	100% { perspective: 1000px; }
}`

// perspectiveCameraStyle keeps the animation running forever. The fixture is
// only ever compared against a fresh layout through an explicit midpoint
// override or through a constant keyframe, so the clock does not affect it.
const perspectiveCameraStyle = `animation-name: camera; animation-duration: 2s;
	animation-timing-function: linear; animation-iteration-count: infinite;
	animation-fill-mode: both;`

// perspectiveCardCSS is the card under the camera. rotateY makes the
// perspective project it, and the offset margins keep the vanishing point away
// from the card's centre.
const perspectiveCardCSS = `#card { width: 100px; height: 50px; margin-left: 45px;
	margin-top: 20px; background-color: rgb(200,20,20); transform: rotateY(45deg); }`

// newPerspectiveCameraPage draws the animated camera fixture.
func newPerspectiveCameraPage(t *testing.T, retain bool) *Page {
	t.Helper()

	source := `<style>body { margin: 0 }` + perspectiveCameraFrames + `
#stage { margin-left: 30px; margin-top: 10px; width: 200px; height: 100px;
	background-color: rgb(10,20,30); ` + perspectiveCameraStyle + ` }
` + perspectiveCardCSS + `
</style><div id="stage"><div id="card"></div></div>`
	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: retain})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// newPerspectiveStaticPage draws the same stage and card with a static
// perspective declaration instead of the animation.
func newPerspectiveStaticPage(t *testing.T, declaration string) *Page {
	t.Helper()

	source := `<style>body { margin: 0 }` + `
#stage { margin-left: 30px; margin-top: 10px; width: 200px; height: 100px;
	background-color: rgb(10,20,30); ` + declaration + ` }
` + perspectiveCardCSS + `
</style><div id="stage"><div id="card"></div></div>`
	p, err := New(Config{HTML: source, Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

// firstElementOp returns the first operation the named element owns.
func firstElementOp(t *testing.T, p *Page, id string) *layout.DisplayOp {
	t.Helper()

	indices := cssmotion.OperationIndices(p.Display(), p.boxByID(id))
	if len(indices) == 0 {
		t.Fatalf("%s owns no operation", id)
	}

	return &p.Display().Ops[indices[0]]
}

// compareProjectedOp compares the painted form of two operations, including
// the display path a 3D projection is finalized into.
func compareProjectedOp(t *testing.T, got, want *layout.DisplayOp, tolerance float64) {
	t.Helper()

	if got.Kind != want.Kind || !nearValue(got.X, want.X, tolerance) ||
		!nearValue(got.Y, want.Y, tolerance) || !nearValue(got.W, want.W, tolerance) ||
		!nearValue(got.H, want.H, tolerance) {
		t.Fatalf("op kind=%v geom=(%v,%v,%v,%v), want kind=%v geom=(%v,%v,%v,%v)",
			got.Kind, got.X, got.Y, got.W, got.H, want.Kind, want.X, want.Y, want.W, want.H)
	}
	gotPath, gotHasPath := got.PathValue()
	wantPath, wantHasPath := want.PathValue()
	if gotHasPath != wantHasPath || len(gotPath.Points) != len(wantPath.Points) {
		t.Fatalf("path set=%v points=%d, want set=%v points=%d",
			gotHasPath, len(gotPath.Points), wantHasPath, len(wantPath.Points))
	}
	for i := range wantPath.Points {
		if !nearValue(gotPath.Points[i].X, wantPath.Points[i].X, tolerance) ||
			!nearValue(gotPath.Points[i].Y, wantPath.Points[i].Y, tolerance) {
			t.Fatalf("path point %d = %+v, want %+v", i, gotPath.Points[i], wantPath.Points[i])
		}
	}
}

func nearValue(got, want, tolerance float64) bool {
	return math.Abs(got-want) <= tolerance
}

// projectedState snapshots the painted form of the named element's first
// operation. The public display shares its operation slice with the retained
// layout state, so a restamp mutates those operations in place; a comparison
// against an earlier frame needs a copy. An affine projection keeps the matrix
// and a projective one is finalized into a display path, so both are recorded.
type projectedState struct {
	x, y, w, h float64
	xformSet   bool
	xform      layout.Matrix2D
	points     []layout.DisplayPoint
}

func newProjectedState(op *layout.DisplayOp) projectedState {
	state := projectedState{
		x: op.X, y: op.Y, w: op.W, h: op.H,
		xformSet: op.XformSet, xform: op.Transform(),
	}
	if path, ok := op.PathValue(); ok {
		state.points = append([]layout.DisplayPoint(nil), path.Points...)
	}

	return state
}

func (s projectedState) same(other projectedState) bool {
	if s.x != other.x || s.y != other.y || s.w != other.w || s.h != other.h ||
		s.xformSet != other.xformSet || s.xform != other.xform ||
		len(s.points) != len(other.points) {
		return false
	}
	for i := range s.points {
		if s.points[i] != other.points[i] {
			return false
		}
	}

	return true
}

func projectedStateOf(t *testing.T, p *Page, id string) projectedState {
	t.Helper()

	return newProjectedState(firstElementOp(t, p, id))
}

func TestPerspectiveRestampMatchesAFreshLayoutAtTheMidpoint(t *testing.T) {
	t.Parallel()

	p := newPerspectiveCameraPage(t, true)
	pointsPerPixel := p.Display().PointsPerPixel

	p.applyRestamp(context.Background(), []motion.Request{{
		ElementID:      "stage",
		Perspective:    750 * pointsPerPixel,
		HasPerspective: true,
	}})
	if got := p.Motion().RestampError; got != "" {
		t.Fatalf("restamp error = %q, want none", got)
	}
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("restamped elements = %d, want 1", got)
	}

	fresh := newPerspectiveStaticPage(t, "perspective: 750px")
	compareProjectedOp(t, firstElementOp(t, p, "card"), firstElementOp(t, fresh, "card"), 1e-6)
}

func TestPerspectiveTickRestampsBetweenFrames(t *testing.T) {
	t.Parallel()

	p := newPerspectiveCameraPage(t, true)
	base := projectedStateOf(t, p, "card")

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := projectedStateOf(t, p, "card")
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("restamped elements = %d, want 1", got)
	}
	if got := p.Motion().RestampError; got != "" {
		t.Fatalf("restamp error = %q, want none", got)
	}
	if first.same(base) {
		t.Fatalf("the first tick kept the pre-animation projection %+v", base)
	}

	time.Sleep(50 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	second := projectedStateOf(t, p, "card")
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("second tick restamped %d elements, want 1", got)
	}
	if second.same(base) {
		t.Fatalf("the second tick kept the pre-animation projection %+v", base)
	}
	if second.same(first) {
		t.Fatalf("the second tick kept the first tick's projection %+v", first)
	}
}

func TestPerspectiveTickMatchesAFreshLayoutAtTheMidpoint(t *testing.T) {
	t.Parallel()

	// A negative delay starts the 10000000 s animation halfway through, so the
	// first tick samples a distance a few millionths of a pixel from 750px
	// even on a loaded machine.
	source := `<style>body { margin: 0 }` + perspectiveCameraFrames + `
#stage { margin-left: 30px; margin-top: 10px; width: 200px; height: 100px;
	background-color: rgb(10,20,30); animation-name: camera;
	animation-duration: 10000000s; animation-delay: -5000000s;
	animation-timing-function: linear; animation-iteration-count: 1;
	animation-fill-mode: both; }
` + perspectiveCardCSS + `
</style><div id="stage"><div id="card"></div></div>`
	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: true})
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
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("restamped elements = %d, want 1", got)
	}

	fresh := newPerspectiveStaticPage(t, "perspective: 750px")
	compareProjectedOp(t, firstElementOp(t, p, "card"), firstElementOp(t, fresh, "card"), 1e-3)
}

func TestPerspectiveOriginTopLeftChangesTheProjection(t *testing.T) {
	t.Parallel()

	const originFrames = `
@keyframes camera {
	0% { perspective: 750px; perspective-origin: top left; }
	100% { perspective: 750px; perspective-origin: top left; }
}`
	source := `<style>body { margin: 0 }` + originFrames + `
#stage { margin-left: 30px; margin-top: 10px; width: 200px; height: 100px;
	background-color: rgb(10,20,30); animation-name: camera;
	animation-duration: 2s; animation-timing-function: linear;
	animation-iteration-count: infinite; animation-fill-mode: both; }
` + perspectiveCardCSS + `
</style><div id="stage"><div id="card"></div></div>`
	p, err := New(Config{HTML: source, Width: 400, Height: 300, RetainState: true})
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
	if got := p.Motion().RestampError; got != "" {
		t.Fatalf("restamp error = %q, want none", got)
	}

	fresh := newPerspectiveStaticPage(t, "perspective: 750px; perspective-origin: top left")
	compareProjectedOp(t, firstElementOp(t, p, "card"), firstElementOp(t, fresh, "card"), 1e-6)

	// The default vanishing point must not produce the same projection, or the
	// comparison above proves nothing about the origin override.
	centred := newPerspectiveStaticPage(t, "perspective: 750px")
	gotPath, _ := firstElementOp(t, p, "card").PathValue()
	centredPath, _ := firstElementOp(t, centred, "card").PathValue()
	if len(gotPath.Points) != len(centredPath.Points) {
		t.Fatalf("path points = %d, want %d", len(gotPath.Points), len(centredPath.Points))
	}
	same := true
	for i := range gotPath.Points {
		if gotPath.Points[i] != centredPath.Points[i] {
			same = false

			break
		}
	}
	if same {
		t.Fatal("the top-left origin painted the same projection as the centred default")
	}
}

func TestPerspectiveOnlyPageEnablesRetentionOnFirstTick(t *testing.T) {
	t.Parallel()

	p := newPerspectiveCameraPage(t, false)
	if !p.Motion().NeedsRetainedState {
		t.Fatal("the bind did not report the perspective keyframe")
	}
	if !p.retainState {
		t.Fatal("the bind did not turn retained layout state on")
	}
	if !p.restampRedraw {
		t.Fatal("the bind did not owe the retained redraw")
	}
	base := projectedStateOf(t, p, "card")
	gen := p.Generation()

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got != 0 {
		t.Fatalf("the first tick restamped %d elements, want the owed redraw instead", got)
	}
	if p.Generation() == gen {
		t.Fatal("the first tick did not redraw with retained layout state")
	}
	if p.restampRedraw {
		t.Fatal("the owed redraw stayed queued after the first tick")
	}
	first := projectedStateOf(t, p, "card")
	if !first.same(base) {
		t.Fatalf("the owed redraw moved the projection from %+v to %+v", base, first)
	}

	time.Sleep(50 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got != 1 {
		t.Fatalf("the second tick restamped %d elements, want 1", got)
	}
	second := projectedStateOf(t, p, "card")
	if second.same(first) {
		t.Fatalf("the second tick kept the first projection %+v, want the camera to move", first)
	}
}

func TestRestampOverrideLeavesPerspectiveOffForATransformRequest(t *testing.T) {
	t.Parallel()

	override := restampOverride(motion.Request{ElementID: "stage", Transform: "rotateY(10deg)"})
	if !override.HasTransform {
		t.Fatal("a transform request did not set HasTransform")
	}
	if override.HasPerspective || override.HasPerspectiveOrigin {
		t.Fatalf("a transform-only request set perspective fields: %+v", override)
	}

	perspectiveOnly := restampOverride(motion.Request{
		ElementID: "stage", Perspective: 375, HasPerspective: true,
	})
	if perspectiveOnly.HasTransform {
		t.Fatal("a perspective-only request replaced the element's static transform")
	}
	if !perspectiveOnly.HasPerspective {
		t.Fatal("a perspective-only request lost the distance")
	}
	if perspectiveOnly.HasPerspectiveOrigin {
		t.Fatal("an undeclared origin was sent to the engine")
	}
}
