package page

import (
	"context"
	"image"
	"testing"
	"time"
)

// newTransitionLayoutPage draws a data-bound class fixture with retained state
// so a geometry transition can splice the display after a redraw.
func newTransitionLayoutPage(t *testing.T, source string, data any) *Page {
	t.Helper()

	p, err := New(Config{HTML: source, Width: 320, Height: 160, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// transitionClassData is the bound class on the fixture's target.
type transitionClassData struct {
	Class string
}

// redrawClass changes the bound class and publishes it.
func redrawClass(t *testing.T, p *Page, data *transitionClassData, class string) {
	t.Helper()
	data.Class = class
	p.SetData(data)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
}

// TestWidthTransitionWidensTheBoxAndMovesASibling is the page-level geometry
// proof: a class change starts a width transition, the relayout grows the box,
// the in-flow sibling moves with it, and the display diff damages the sibling.
func TestWidthTransitionWidensTheBoxAndMovesASibling(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#row { display: flex; width: 300px; height: 60px; }
#a { width: 40px; height: 20px; background-color: rgb(20,40,60);
	transition-property: width; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { width: 120px; }
#b { width: 40px; height: 20px; background-color: rgb(80,100,120); }
</style><div id="row"><div id="a" class="{{.Class}}"></div><div id="b"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	start := time.Now()
	redrawClass(t, p, data, "on")

	if got := p.Motion().TransitionsBound; got != 1 {
		t.Fatalf("bound transitions = %d, want the width run", got)
	}
	p.TakeDirty()
	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))

	a, b := p.boxByID("a"), p.boxByID("b")
	if a.W <= 40 || a.W >= 120 {
		t.Fatalf("midpoint width = %v, want between 40 and 120", a.W)
	}
	if b.X <= 40 {
		t.Fatalf("sibling X = %v, want it past the base 40", b.X)
	}
	if got := p.Motion().LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts = %d, want 1", got)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("relayout error = %q, want none", got)
	}

	want := image.Rect(int(b.X), int(b.Y), int(b.X+b.W), int(b.Y+b.H))
	rects, ok := p.TakeDirtyRects()
	if !ok {
		t.Fatal("the width transition marked no damage")
	}
	covered := false
	for _, rect := range rects {
		if rect.Min.X <= want.Min.X && rect.Min.Y <= want.Min.Y &&
			rect.Max.X >= want.Max.X && rect.Max.Y >= want.Max.Y {
			covered = true
		}
	}
	if !covered {
		t.Fatalf("damage %v does not cover the moved sibling %v", rects, want)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	if got := p.boxByID("a").W; got != 120 {
		t.Fatalf("final width = %v, want the destination 120", got)
	}
	if got := p.boxByID("b").X; got != 120 {
		t.Fatalf("final sibling X = %v, want 120 beside the final width", got)
	}
}

// TestHeightTransitionGrowsTheBox checks the vertical extent through the same
// relayout path. The fixed-height row keeps the change inside the patch.
func TestHeightTransitionGrowsTheBox(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#row { width: 200px; height: 120px; }
#a { height: 20px; background-color: rgb(20,40,60);
	transition-property: height; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { height: 80px; }
</style><div id="row"><div id="a" class="{{.Class}}"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	start := time.Now()
	redrawClass(t, p, data, "on")

	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))
	if got := p.boxByID("a").H; got <= 20 || got >= 80 {
		t.Fatalf("midpoint height = %v, want between 20 and 80", got)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("height relayout error = %q, want none", got)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	if got := p.boxByID("a").H; got != 80 {
		t.Fatalf("final height = %v, want 80", got)
	}
}

// TestLeftTopTransitionMovesTheBox checks the inset path: the position change
// becomes a relayout override and the element's border box follows it.
func TestLeftTopTransitionMovesTheBox(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#stage { position: relative; width: 300px; height: 150px; }
#a { position: absolute; left: 0; top: 0; width: 40px; height: 20px;
	background-color: rgb(20,40,60);
	transition-property: left, top; transition-duration: 400ms;
	transition-timing-function: linear; }
#a.on { left: 100px; top: 60px; }
</style><div id="stage"><div id="a" class="{{.Class}}"></div></div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	start := time.Now()
	redrawClass(t, p, data, "on")

	if got := p.Motion().TransitionsBound; got != 2 {
		t.Fatalf("bound transitions = %d, want left and top", got)
	}
	p.stepLayoutMotion(context.Background(), start.Add(200*time.Millisecond))
	box := p.boxByID("a")
	if box.X <= 0 || box.X >= 100 {
		t.Fatalf("midpoint left = %v, want between 0 and 100", box.X)
	}
	if box.Y <= 0 || box.Y >= 60 {
		t.Fatalf("midpoint top = %v, want between 0 and 60", box.Y)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("inset relayout error = %q, want none", got)
	}

	p.stepLayoutMotion(context.Background(), start.Add(600*time.Millisecond))
	box = p.boxByID("a")
	if box.X != 100 || box.Y != 60 {
		t.Fatalf("final box = (%v,%v), want (100,60)", box.X, box.Y)
	}
}

// TestVisibilityTransitionHidesTheRemainingFill checks the presence path at the
// page seam: a leaving element keeps the fill it still owns visible for the run
// and hides it at the end.
func TestVisibilityTransitionHidesTheRemainingFill(t *testing.T) {
	t.Parallel()

	source := `<style>
body { margin: 0 }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	transition-property: visibility; transition-duration: 400ms;
	transition-behavior: allow-discrete; transition-timing-function: linear; }
#target.on { visibility: hidden; }
</style><div id="target" class="{{.Class}}">hidden words</div>`
	data := &transitionClassData{}
	p := newTransitionLayoutPage(t, source, data)
	start := time.Now()
	redrawClass(t, p, data, "on")

	box := p.boxByID("target")
	if len(box.OperationIDs) == 0 {
		t.Fatal("the hidden element kept no operation to fade")
	}
	op := transitionWireOp(t, p)
	if op.Alpha != 1 {
		t.Fatalf("alpha at the bind = %v, want the visible 1", op.Alpha)
	}
	p.setMotionView()
	p.motion.Step(p.draw.Display, start.Add(200*time.Millisecond))
	op = transitionWireOp(t, p)
	if op.Alpha != 1 {
		t.Fatalf("alpha mid-run = %v, want the visible 1", op.Alpha)
	}

	p.motion.Step(p.draw.Display, start.Add(600*time.Millisecond))
	op = transitionWireOp(t, p)
	if op.Alpha > 0.01 {
		t.Fatalf("alpha at the end = %v, want the hidden endpoint; stats %+v",
			op.Alpha, p.Motion())
	}
}
