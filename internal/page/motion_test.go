package page

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"image"
	"math"
	"strings"
	"testing"
	"time"
)

// layoutWidenFrames animates a box's width, the cheapest layout-tier property
// to inspect: the element's own border box changes and an in-flow sibling
// moves.
const layoutWidenFrames = `
@keyframes widen {
	0% { width: 40px; }
	100% { width: 120px; }
}`

// layoutSlideFrames moves a neighbour through the 2D matrix path, so a layout
// patch must preserve its clock while it relayouts the row.
const layoutSlideFrames = `
@keyframes scoot {
	0% { transform: translateX(0px); }
	100% { transform: translateX(90px); }
}`

const layoutWidenStyle = `animation-name: widen; animation-duration: 200ms;
	animation-timing-function: linear; animation-fill-mode: both;`

const layoutSlideStyle = `animation-name: scoot; animation-duration: 200ms;
	animation-timing-function: linear; animation-fill-mode: both;`

// newLayoutPage draws a flex row with one width-animating box and a sibling,
// and drains the first-frame damage.
func newLayoutPage(t *testing.T, retain bool) *Page {
	t.Helper()

	source := `<style>` + layoutWidenFrames + `
#row { display: flex; }
#grow { width: 40px; height: 20px; background-color: rgb(20,40,60); ` +
		layoutWidenStyle + ` }
#next { width: 30px; height: 20px; background-color: rgb(80,100,120); }
</style><div id="row"><div id="grow"></div><div id="next"></div></div>`

	p, err := New(Config{HTML: source, Width: 300, Height: 100, RetainState: retain})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// tickLayout drives one tick after a short sleep, so a 200ms linear animation
// moves a visible amount.
func tickLayout(t *testing.T, p *Page) {
	t.Helper()

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// firstOpOf returns the first display operation the element with id owns.
func firstOpOf(t *testing.T, p *Page, id string) int {
	t.Helper()

	indices := motion.OperationIndices(p.Display(), boxByID(p, id))
	if len(indices) == 0 {
		t.Fatalf("%s owns no operation", id)
	}

	return indices[0]
}

// textOpOf returns the first text operation the element with id owns, or nil.
func textOpOf(p *Page, id string) *DisplayOp {
	d := p.Display()
	for _, index := range motion.OperationIndices(d, boxByID(p, id)) {
		if d.Ops[index].Kind == layout.DisplayOpText {
			return &d.Ops[index]
		}
	}

	return nil
}

// TestLayoutWidthKeyframeWidensTheBoxAndMovesASibling is the phase 04
// acceptance proof: a width keyframe reaches the engine's override relayout
// path, the element's border box grows, and the in-flow sibling moves with it.
func TestLayoutWidthKeyframeWidensTheBoxAndMovesASibling(t *testing.T) {
	t.Parallel()

	p := newLayoutPage(t, true)
	before := boxByID(p, "grow")
	nextBefore := boxByID(p, "next")

	tickLayout(t, p)

	after := boxByID(p, "grow")
	nextAfter := boxByID(p, "next")
	if after.W <= before.W {
		t.Fatalf("grow width = %v, want it past the base %v", after.W, before.W)
	}
	if nextAfter.X <= nextBefore.X {
		t.Fatalf("sibling X = %v, want it past the base %v", nextAfter.X, nextBefore.X)
	}
	if got := p.Motion().LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts = %d, want 1", got)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("relayout error = %q, want none", got)
	}

	// The relayout rebinds the element's own player too, so the next frame has
	// to continue past the first one instead of restarting at the base width.
	tickLayout(t, p)
	if again := boxByID(p, "grow").W; again < after.W {
		t.Fatalf("grow width = %v on the second tick, want at least the first %v", again, after.W)
	}
}

// TestLayoutRelayoutDoesNotRestartANeighbourTransform pins the phase 01 rebind
// contract on the new adoption path: a splice bumps the generation and rebinds
// motion, but the neighbour's transform player keeps its clock.
func TestLayoutRelayoutDoesNotRestartANeighbourTransform(t *testing.T) {
	t.Parallel()

	source := `<style>` + layoutWidenFrames + layoutSlideFrames + `
#row { display: flex; }
#grow { width: 40px; height: 20px; ` + layoutWidenStyle + ` }
#scoot { width: 30px; height: 20px; background-color: rgb(80,100,120); ` +
		layoutSlideStyle + ` }
</style><div id="row"><div id="grow"></div><div id="scoot"></div></div>`

	p, err := New(Config{HTML: source, Width: 320, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	first := motionTargetX(p, firstOpOf(t, p, "scoot"))
	if first <= 0 {
		t.Fatalf("the neighbour transform did not advance: X = %v", first)
	}

	// The tick that relayouts the width box must leave the neighbour's
	// translation at or past where it was, not restart it at zero.
	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := motionTargetX(p, firstOpOf(t, p, "scoot")); got < first {
		t.Fatalf("the neighbour transform restarted: X = %v, want at least %v", got, first)
	}
	if got := p.Motion().LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts = %d, want 1", got)
	}
	if got := p.Motion().Rebinds; got == 0 {
		t.Fatal("the relayout did not rebind the paint players")
	}
}

// TestLayoutBudgetRunsOneElementPerFrame proves the per-frame budget: the
// second width-animating element waits, is reported as queued, and samples on
// the frame that runs it.
func TestLayoutBudgetRunsOneElementPerFrame(t *testing.T) {
	t.Parallel()

	source := `<style>` + layoutWidenFrames + `
#row { display: flex; }
.card { width: 40px; height: 20px; ` + layoutWidenStyle + ` }
</style><div id="row"><div id="a" class="card"></div><div id="b" class="card"></div></div>`

	p, err := New(Config{HTML: source, Width: 320, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	aBefore, bBefore := boxByID(p, "a").W, boxByID(p, "b").W

	tickLayout(t, p)
	stats := p.Motion()
	if got := stats.LayoutBound; got != 2 {
		t.Fatalf("bound layout elements = %d, want 2", got)
	}
	if got := stats.LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts in one frame = %d, want the budget of 1", got)
	}
	if !stats.LayoutBudgetHit || stats.LayoutQueued == 0 {
		t.Fatalf("budget hit = %v, queued = %d, want the second element waiting",
			stats.LayoutBudgetHit, stats.LayoutQueued)
	}

	aFirst, bFirst := boxByID(p, "a").W, boxByID(p, "b").W
	switch {
	case aFirst == aBefore && bFirst == bBefore:
		t.Fatal("neither element moved on the first frame")
	case aFirst != aBefore && bFirst != bBefore:
		t.Fatal("both elements relaid out in one frame, so the budget was exceeded")
	}

	tickLayout(t, p)
	if aFirst == aBefore {
		if got := boxByID(p, "a").W; got == aFirst {
			t.Fatalf("the queued first element never ran: width %v", got)
		}
	}
	if bFirst == bBefore {
		if got := boxByID(p, "b").W; got == bFirst {
			t.Fatalf("the queued second element never ran: width %v", got)
		}
	}
}

// TestRetainStateTurnsOnOnlyForALayoutKeyframe pins the on-demand memory
// switch: a 2D transform page stays unretained, and a width page pays for the
// retained layout state so the engine can patch it.
func TestRetainStateTurnsOnOnlyForALayoutKeyframe(t *testing.T) {
	t.Parallel()

	slide := newMotionPage(t, motionRunning)
	if slide.retainState {
		t.Fatal("a 2D transform page turned retained layout state on")
	}
	if got := slide.Motion().NeedsRetainedState; got {
		t.Fatal("a 2D transform page asked for retained layout state")
	}
	if got := slide.Motion().LayoutBound; got != 0 {
		t.Fatalf("a 2D transform page bound %d layout players", got)
	}

	widen := newLayoutPage(t, false)
	if !widen.retainState {
		t.Fatal("the width page did not turn retained layout state on")
	}
	if !widen.Motion().NeedsRetainedState {
		t.Fatal("the width page did not report the layout keyframe")
	}
	if got := widen.Motion().LayoutBound; got != 1 {
		t.Fatalf("layout players = %d, want 1", got)
	}
	if !widen.restampRedraw {
		t.Fatal("the width page did not owe the redraw that retains layout state")
	}
}

// TestLayoutFirstTickPaysTheRetainedStateRedraw checks the on-demand switch
// end to end: the frame that discovers the layout keyframe owes one redraw, and
// the next tick applies the first sample. The pose is not dropped in between.
func TestLayoutFirstTickPaysTheRetainedStateRedraw(t *testing.T) {
	t.Parallel()

	p := newLayoutPage(t, false)
	gen := p.Generation()

	tickLayout(t, p)
	if got := boxByID(p, "grow").W; got != 40 {
		t.Fatalf("grow width = %v on the redraw frame, want the base 40", got)
	}
	if p.restampRedraw {
		t.Fatal("the owed retained-state redraw stayed queued")
	}
	if p.Generation() == gen {
		t.Fatal("the first tick did not redraw with retained layout state")
	}

	tickLayout(t, p)
	if got := boxByID(p, "grow").W; got <= 40 {
		t.Fatalf("grow width = %v on the frame after the redraw, want it past the base", got)
	}
	if got := p.Motion().LayoutRelayouts; got != 1 {
		t.Fatalf("relayouts = %d, want 1", got)
	}
}

// TestTwoDimensionalPageStaysOffTheRelayoutPath checks that the driver never
// reaches the engine pass without a layout keyframe: the generation stays put
// and no relayout is counted.
func TestTwoDimensionalPageStaysOffTheRelayoutPath(t *testing.T) {
	t.Parallel()

	p := newMotionPage(t, motionRunning)
	gen := p.Generation()

	for i := 0; i < 3; i++ {
		tickLayout(t, p)
	}

	if p.Generation() != gen {
		t.Fatal("a 2D page rebuilt its display, so it reached the relayout path")
	}
	if got := p.Motion().LayoutRelayouts; got != 0 {
		t.Fatalf("a 2D page relaid out %d elements", got)
	}
	if got := p.Motion().LayoutBound; got != 0 {
		t.Fatalf("a 2D page bound %d layout players", got)
	}
}

// TestMalformedLayoutValueFailsTheBind keeps a bad width from silently
// animating the base value: the bind records the parse error once.
func TestMalformedLayoutValueFailsTheBind(t *testing.T) {
	t.Parallel()

	source := `<style>
@keyframes broken {
	0% { width: 40px; }
	100% { width: banana; }
}
#grow { width: 40px; height: 20px; animation-name: broken;
	animation-duration: 60ms; animation-fill-mode: both; }
</style><div id="grow"></div>`

	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	stats := p.Motion()
	if !strings.Contains(stats.LastError, "invalid layout value") {
		t.Fatalf("bind error = %q, want the invalid layout value", stats.LastError)
	}
	if got := stats.LayoutBound; got != 0 {
		t.Fatalf("layout players = %d, want the failed box left out", got)
	}
}

// TestLayoutRelayoutDamagesTheMovedSibling checks the tile damage rule: a
// width change moves operations outside the patched element's own box, and the
// display diff has to cover them.
func TestLayoutRelayoutDamagesTheMovedSibling(t *testing.T) {
	t.Parallel()

	p := newLayoutPage(t, true)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
	tickLayout(t, p)

	next := boxByID(p, "next")
	want := image.Rect(
		int(next.X), int(next.Y), int(next.X+next.W), int(next.Y+next.H))
	rects, ok := p.TakeDirtyRects()
	if !ok {
		t.Fatal("the relayout marked no damage")
	}
	if p.LastPatch().Full {
		t.Fatal("the relayout dirtied the whole page instead of the display diff")
	}
	for _, rect := range rects {
		if rect.Min.X <= want.Min.X && rect.Min.Y <= want.Min.Y &&
			rect.Max.X >= want.Max.X && rect.Max.Y >= want.Max.Y {
			return
		}
	}
	t.Fatalf("damage %v does not cover the moved sibling %v", rects, want)
}

// TestQueuedLayoutSamplesAtTheFrameItRuns pins the sample time of a queued
// element: the request it applies is the value at that frame's clock, not the
// value when it queued. The two-element row queues b behind a on the first
// frame; b's own width at that later frame is further along than the first
// frame's would have been.
func TestQueuedLayoutSamplesAtTheFrameItRuns(t *testing.T) {
	t.Parallel()

	source := `<style>` + layoutWidenFrames + `
#row { display: flex; }
.card { width: 40px; height: 20px; ` + layoutWidenStyle + ` }
</style><div id="row"><div id="a" class="card"></div><div id="b" class="card"></div></div>`

	p, err := New(Config{HTML: source, Width: 320, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	// The queue is ordered by element id, so b is the deferred element.
	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := boxByID(p, "b").W; got != 40 {
		t.Fatalf("b width = %v on the frame it waited, want the base 40", got)
	}

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := boxByID(p, "b").W; got <= 40 {
		t.Fatalf("b width = %v on the frame it ran, want it past the base", got)
	}
}

// TestReturnLayoutKeepsAFailedSampleQueued checks the retained-state debt path:
// a request the page cannot apply yet goes back to the head of the queue, so a
// finished animation is not dropped.
func TestReturnLayoutKeepsAFailedSampleQueued(t *testing.T) {
	t.Parallel()

	rt := motion.New()
	request := motion.LayoutRequest{ElementID: "grow"}
	rt.ReturnLayout([]motion.LayoutRequest{request})

	if got := rt.Stats().LayoutQueued; got != 1 {
		t.Fatalf("queued = %d, want the returned element", got)
	}
}

// TestLayoutReflowEscapeKeepsTheDisplay pins the engine limit the phase cannot
// hide: a height override that resizes an auto-height containing block is
// refused with ErrOverrideReflow, the page keeps drawing its last good frame,
// and the failure is reported instead of leaving a half-laid-out page.
func TestLayoutReflowEscapeKeepsTheDisplay(t *testing.T) {
	t.Parallel()

	source := `<style>
@keyframes taller {
	0% { height: 20px; }
	100% { height: 60px; }
}
#grow { width: 40px; height: 20px; animation-name: taller;
	animation-duration: 60ms; animation-timing-function: linear; animation-fill-mode: both; }
</style><div id="row"><div id="grow"></div></div>`

	p, err := New(Config{HTML: source, Width: 240, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
	before := p.Display()
	gen := p.Generation()

	tickLayout(t, p)

	if p.Display() != before {
		t.Fatal("a refused reflow replaced the display")
	}
	if p.Generation() != gen {
		t.Fatalf("a refused reflow bumped the generation to %d", p.Generation())
	}
	if got := p.Motion().LayoutRelayouts; got != 0 {
		t.Fatalf("relayouts after the refusal = %d, want 0", got)
	}
	if got := p.Motion().LayoutError; !strings.Contains(got, "escapes its containing block") {
		t.Fatalf("layout error = %q, want the reflow refusal", got)
	}
}

// TestLayoutFontSizeAnimationReflowsText checks an inherited layout property:
// the sampled font-size reaches the text run through the engine's override
// pass, so the shaped text grows without a full-page redraw.
func TestLayoutFontSizeAnimationReflowsText(t *testing.T) {
	t.Parallel()

	source := `<style>
@keyframes big {
	0% { font-size: 10px; }
	100% { font-size: 30px; }
}
#box { width: 200px; height: 80px; font-size: 10px; animation-name: big;
	animation-duration: 200ms; animation-timing-function: linear; animation-fill-mode: both; }
</style><div id="box">hello world</div>`

	p, err := New(Config{HTML: source, Width: 260, Height: 100, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	before := textOpOf(p, "box")
	if before == nil {
		t.Fatal("the fixture painted no text")
	}
	beforeWidth := before.W

	tickLayout(t, p)

	after := textOpOf(p, "box")
	if after == nil {
		t.Fatal("the relayout dropped the text")
	}
	if after.W <= beforeWidth {
		t.Fatalf("text width = %v, want it past the base %v", after.W, beforeWidth)
	}
	if got := p.Motion().LayoutError; got != "" {
		t.Fatalf("font-size relayout error = %q, want none", got)
	}
}

// TestLayoutTierOverridesAreAccepted walks the parsed property tier through a
// real engine relayout. The parser test pins the value grammar and this one
// pins that the string the driver emits is a value the engine applies, which is
// the seam between the two packages.
func TestLayoutTierOverridesAreAccepted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		property string
		value    string
	}{
		{"width", "width", "80px"},
		{"height", "height", "40px"},
		{"min width", "min-width", "20px"},
		{"max width", "max-width", "120px"},
		{"min height", "min-height", "10px"},
		{"max height", "max-height", "50px"},
		{"inset", "inset", "2px"},
		{"top", "top", "2px"},
		{"logical inset", "inset-inline-start", "5px"},
		{"margin", "margin", "4px"},
		{"logical margin", "margin-block-start", "3px"},
		{"padding", "padding", "6px"},
		{"gap", "gap", "10px"},
		{"flex basis", "flex-basis", "30px"},
		{"border width", "border-top-width", "3px"},
		{"border radius", "border-radius", "8px"},
		{"corner radius", "border-top-left-radius", "4px"},
		{"font size", "font-size", "20px"},
		{"line height", "line-height", "1.5"},
		{"letter spacing", "letter-spacing", "1px"},
		{"word spacing", "word-spacing", "2px"},
		{"percent width", "width", "50%"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			source := `<div id="row" style="display:flex;width:220px;height:80px">` +
				`<div id="target" style="display:flex;width:40px;height:20px;background-color:#224466">` +
				`<span>text</span></div><div style="width:20px;height:20px"></div></div>`
			p, err := New(Config{HTML: source, Width: 240, Height: 100, RetainState: true})
			if err != nil {
				t.Fatal(err)
			}
			if err := p.Redraw(context.Background()); err != nil {
				t.Fatal(err)
			}

			err = applyLayoutRequest(p, context.Background(), motion.LayoutRequest{
				ElementID: "target",
				Overrides: []motion.LayoutOverride{{Property: test.property, Value: test.value}},
			})
			if err != nil {
				t.Fatalf("%s: %v: %s", test.property, err, test.value)
			}
		})
	}
}

func TestTickManualOverlapDiagnostic(t *testing.T) {
	html := `<style>
@keyframes pulse { from { opacity: 0.2; } to { opacity: 1; } }
#anim { width: 40px; height: 40px; background: #ff0000; animation: pulse 2s linear infinite; }
#other { width: 40px; height: 40px; background: #0000ff; }
</style><div id="anim"></div><div id="other"></div>`

	p, err := New(Config{HTML: html, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	// Mutating an element owned by an active CSS animation triggers ManualOverlap.
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].R > 0.5 && d.Ops[i].B < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	st := p.Motion()
	if st.ManualOverlap != 1 || st.ManualOverlapID != "anim" {
		t.Fatalf("after overlapping tick: ManualOverlap=%d ID=%q, want 1 / %q", st.ManualOverlap, st.ManualOverlapID, "anim")
	}

	// Mutating a non-animated element does not increment ManualOverlap.
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].B > 0.5 && d.Ops[i].R < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().ManualOverlap; got != 1 {
		t.Fatalf("non-overlapping tick changed ManualOverlap to %d, want 1", got)
	}

	// Turning animations off prevents overlap diagnostics when driving #anim manually.
	p.SetAnimations(false)
	p.SetTick(func(context.Context) error {
		d := p.Display()
		for i := range d.Ops {
			if d.Ops[i].R > 0.5 && d.Ops[i].B < 0.5 {
				d.Ops[i].X += 5
				break
			}
		}
		return nil
	})
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().ManualOverlap; got != 1 {
		t.Fatalf("after SetAnimations(false), ManualOverlap=%d, want 1 (unchanged)", got)
	}
}

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

	indices := motion.OperationIndices(p.Display(), boxByID(p, id))
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

	applyRestamp(p, context.Background(), []motion.Request{{
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

// TestRestampReusesBoundPlayers checks that a re-projection pass keeps the
// operation slice in place, so the page does not rebuild every animation player
// on each animated frame.
func TestRestampReusesBoundPlayers(t *testing.T) {
	t.Parallel()

	p := newRestampPage(t, true)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got == 0 {
		t.Fatal("the first tick did not restamp")
	}
	before := p.Motion().Rebinds
	layerGen := p.LayerGeneration()

	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.Motion().Restamped; got == 0 {
		t.Fatal("the second tick did not restamp")
	}
	if got := p.Motion().Rebinds; got != before {
		t.Fatalf("rebinds = %d, want %d; the restamp rebuilt the players", got, before)
	}
	if got := p.LayerGeneration(); got != layerGen {
		t.Fatalf("layer generation = %d, want %d; a restamp must reuse layer textures", got, layerGen)
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
		indices := motion.OperationIndices(p.Display(), boxByID(p, id))
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
	adoptRestamp(p, &layout.ElementPatch{Display: &clone, Old: box, New: box})

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
	adoptRestamp(p, &layout.ElementPatch{Display: &elsewhere,
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
	adoptRestamp(p, &layout.ElementPatch{Display: &clone,
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
	applyRestamp(p, context.Background(), []motion.Request{{
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

	applyRestamp(p, context.Background(), []motion.Request{{
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

// pageTimelineHTML declares the page scroller under a name and animates one
// target from it. The tall row gives the page a scroll range.
const pageTimelineHTML = `<style>body { margin: 0 }` + motionSlideFrames + `
#page { scroll-timeline: --page y; }
#target { width: 40px; height: 20px; background-color: rgb(20,40,60);
	animation-name: slide; animation-duration: 1s;
	animation-timing-function: linear; animation-fill-mode: both;
	animation-timeline: --page; }
#tall { height: 500px; }
</style><div id="page"><div id="target"></div><div id="tall"></div></div>`

// newPageTimelinePage draws the named scroll-timeline fixture.
func newPageTimelinePage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{HTML: pageTimelineHTML, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// pageScrollSpan is the page scroller range the fixture produces.
func pageScrollSpan(p *Page) float64 {
	content := contentRect(p)

	return float64(content.Max.Y - p.height)
}

// TestPageNamedScrollTimelineFollowsOffset checks the page wiring: a named
// timeline declared on an ancestor resolves from the box metadata, samples
// the window's scroll offset, and paints the right phase.
func TestPageNamedScrollTimelineFollowsOffset(t *testing.T) {
	p := newPageTimelinePage(t)
	index := motionTargetIndex(t, p)
	base := motionTargetX(p, index)

	var scroller Box
	for _, box := range p.Boxes() {
		if box.ID == "page" {
			scroller = box
		}
	}
	if len(scroller.Timelines) == 0 || scroller.Timelines[0].Name != "--page" {
		t.Fatalf("the engine exported no --page timeline on #page: %+v", scroller.Timelines)
	}

	span := pageScrollSpan(p)
	if span <= 0 {
		t.Fatalf("the fixture page does not scroll: span %v", span)
	}
	scale := p.Display().PointsPerPixel
	if scale <= 0 {
		scale = 1
	}
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got) > 0.01 {
		t.Fatalf("top sample moved %v pixels, want 0", got)
	}

	p.SetScrollOffset(0, int(span/2))
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got-60) > 0.5 {
		t.Fatalf("middle sample moved %v pixels, want 60", got)
	}

	p.SetScrollOffset(0, int(span))
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := (motionTargetX(p, index) - base) / scale; math.Abs(got-120) > 0.5 {
		t.Fatalf("bottom sample moved %v pixels, want 120", got)
	}
}

// TestPageScrollTimelineDoesNotTick checks pacing: a still page with a scroll
// animation samples once at the offset and never asks for another frame.
func TestPageScrollTimelineDoesNotTick(t *testing.T) {
	p := newPageTimelinePage(t)
	p.SetScrollOffset(0, 100)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.NextFrameAt()
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.NextFrameAt(); !got.IsZero() {
		t.Fatalf("a still scroll animation asked for a frame at %v", got)
	}
}

// triggerFrames is the transform the trigger fixtures animate.
const triggerFrames = `
@keyframes trigger-slide {
	0% { transform: translateX(0px); }
	100% { transform: translateX(100px); }
}`

// triggeredTargetCSS declares the animated element. It follows the page scroll
// timeline so the sampled pose is a pure function of the offset: the tests
// never rely on wall-clock time.
const triggeredTargetCSS = `#target { width: 20px; height: 20px;
	background-color: rgb(10,20,30);
	animation-name: trigger-slide; animation-duration: 1s;
	animation-timing-function: linear; animation-fill-mode: both;
	animation-timeline: scroll();
	animation-trigger: --play play --pause pause --reset reset; }`

// newTriggerPage draws a page with three trigger sources above one target and
// a tall row that gives the page a scroll range.
func newTriggerPage(t *testing.T) *Page {
	t.Helper()

	source := `<style>body { margin: 0 }` + triggerFrames + `
#play, #pause, #reset { width: 20px; height: 20px; }
#play { animation-trigger: --play play; }
#pause { animation-trigger: --pause pause; }
#reset { animation-trigger: --reset reset; }
` + triggeredTargetCSS + `
#tall { height: 500px; }
</style><div id="play"></div><div id="pause"></div><div id="reset"></div>
<div id="target"></div><div id="tall"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	return p
}

// boxTranslate is the first operation's animation translate the box owns, in
// CSS pixels. The fixtures have no static transform, so the matrix E is the
// animation's own offset.
func boxTranslate(t *testing.T, p *Page, id string) float64 {
	t.Helper()

	box := boxByID(p, id)
	if box.ID == "" {
		t.Fatalf("box %q is missing", id)
	}
	indices := motion.OperationIndices(p.Display(), box)
	if len(indices) == 0 {
		t.Fatalf("box %q owns no operation", id)
	}
	op := &p.Display().Ops[indices[0]]
	scale := p.Display().PointsPerPixel
	if scale <= 0 {
		scale = 1
	}

	return op.Transform().E / scale
}

// clickCenter clicks the middle of the box with id. The tests reset the scroll
// offset before every click, so the content point is the viewport point.
func clickCenter(t *testing.T, p *Page, id string) {
	t.Helper()

	box := boxByID(p, id)
	if box.ID == "" {
		t.Fatalf("box %q is missing", id)
	}
	if err := p.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("click %q: %v", id, err)
	}
}

// tickAtOffset scrolls the page and samples once.
func tickAtOffset(t *testing.T, p *Page, offset int) {
	t.Helper()

	p.SetScrollOffset(0, offset)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// TestClickTriggerPlaysPausesAndResets drives the same animation with three
// clicks: play starts it, pause freezes the current timeline phase, and reset
// returns it to the start. Each source element declares the name it raises and
// the target listens for all three names.
func TestClickTriggerPlaysPausesAndResets(t *testing.T) {
	p := newTriggerPage(t)
	span := pageScrollSpan(p)
	if span <= 0 {
		t.Fatalf("fixture page does not scroll: %v", span)
	}

	// Pending trigger: the target sits at its base paint.
	tickAtOffset(t, p, int(span/2))
	if got := boxTranslate(t, p, "target"); got != 0 {
		t.Fatalf("pending trigger moved %v pixels, want 0", got)
	}

	clickCenter(t, p, "play")
	tickAtOffset(t, p, int(span/2))
	played := boxTranslate(t, p, "target")
	if math.Abs(played-50) > 0.5 {
		t.Fatalf("play sample = %v pixels, want 50", played)
	}

	p.SetScrollOffset(0, 0)
	clickCenter(t, p, "pause")
	tickAtOffset(t, p, int(span*4/5))
	if got := boxTranslate(t, p, "target"); math.Abs(got-played) > 0.5 {
		t.Fatalf("pause sample = %v pixels, want the frozen %v", got, played)
	}

	p.SetScrollOffset(0, 0)
	clickCenter(t, p, "reset")
	tickAtOffset(t, p, int(span*4/5))
	if got := boxTranslate(t, p, "target"); math.Abs(got) > 0.5 {
		t.Fatalf("reset sample = %v pixels, want 0", got)
	}
}

// TestClickAppliesEveryTriggerNameInOrder clicks the animated element, which
// declares play, pause, and reset on one item. The actions apply in
// declaration order, so the last one (reset) owns the sampled pose.
func TestClickAppliesEveryTriggerNameInOrder(t *testing.T) {
	p := newTriggerPage(t)
	span := pageScrollSpan(p)

	clickCenter(t, p, "target")
	tickAtOffset(t, p, int(span/2))

	if got := boxTranslate(t, p, "target"); math.Abs(got) > 0.5 {
		t.Fatalf("ordered actions left %v pixels, want the reset pose 0", got)
	}
}

// TestTriggerScopeLimitsDispatch puts one listener inside the scope that names
// the event and one outside. A click on the source fires only the inside one.
func TestTriggerScopeLimitsDispatch(t *testing.T) {
	source := `<style>body { margin: 0 }` + triggerFrames + `
#scope { trigger-scope: --go; }
#go { width: 20px; height: 20px; }
.inside, .outside { width: 20px; height: 20px; background-color: rgb(10,20,30);
	animation-name: trigger-slide; animation-duration: 1s;
	animation-timing-function: linear; animation-fill-mode: both;
	animation-timeline: scroll(); animation-trigger: --go play; }
#tall { height: 500px; }
</style><section id="scope"><div id="go"></div>
<div id="inside" class="inside"></div></section>
<div id="outside" class="outside"></div><div id="tall"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	if err := p.Click(context.Background(), 10, 10); err != nil {
		t.Fatal(err)
	}
	tickAtOffset(t, p, int(pageScrollSpan(p)/2))

	if got := boxTranslate(t, p, "inside"); math.Abs(got-50) > 0.5 {
		t.Fatalf("inside sample = %v pixels, want 50", got)
	}
	if got := boxTranslate(t, p, "outside"); got != 0 {
		t.Fatalf("outside sample moved %v pixels, want 0", got)
	}
}

// TestKeyboardActivationFiresTrigger activates a focused button with Enter.
func TestKeyboardActivationFiresTrigger(t *testing.T) {
	source := `<style>body { margin: 0 }` + triggerFrames + `
#play { width: 20px; height: 20px; animation-trigger: --play play; }
` + triggeredTargetCSS + `
#tall { height: 500px; }
</style><button id="play"></button><div id="target"></div><div id="tall"></div>`
	p, err := New(Config{HTML: source, Width: 200, Height: 100})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()

	if err := p.Focus(ctx, "play"); err != nil {
		t.Fatalf("focus: %v", err)
	}
	if got := formState(p).FocusID; got != "play" {
		t.Fatalf("focused id = %q, want play", got)
	}
	if err := p.KeyDown(ctx, "enter"); err != nil {
		t.Fatalf("key: %v", err)
	}
	tickAtOffset(t, p, int(pageScrollSpan(p)/2))

	if got := boxTranslate(t, p, "target"); math.Abs(got-50) > 0.5 {
		t.Fatalf("keyboard activation sample = %v pixels, want 50", got)
	}
}

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

	indices := motion.OperationIndices(d, target)
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

	// A real input change rebuilds the display and rebinds the players.
	gen := p.Generation()
	p.SetSize(220, 100)
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Generation() == gen {
		t.Fatal("a resized Redraw did not bump the generation")
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

// TestIdenticalRedrawKeepsAnimatedFrame pins frame reuse during an active
// animation: Tick already sampled the pose into the committed operations, so
// a Redraw with the same input keeps that display, its pose, and the players'
// phase instead of laying the page out again.
func TestIdenticalRedrawKeepsAnimatedFrame(t *testing.T) {
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
	if p.Motion().Active == 0 {
		t.Fatal("the fixture animation is not active")
	}
	before := motionTargetX(p, index)
	if before <= 0 {
		t.Fatalf("the animation did not advance before the redraw: X = %v", before)
	}
	gen, layouts, rebinds := p.Generation(), p.stats.fullLayouts, p.Motion().Rebinds
	display := p.Display()

	for range 3 {
		if err := p.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if p.Generation() != gen || p.stats.fullLayouts != layouts || p.Display() != display {
		t.Fatalf("identical Redraw rebuilt the frame: generation %d->%d, layouts %d->%d",
			gen, p.Generation(), layouts, p.stats.fullLayouts)
	}
	if got := p.Motion().Rebinds; got != rebinds {
		t.Fatalf("identical Redraw rebound the players: rebinds %d->%d", rebinds, got)
	}
	if got := motionTargetX(p, index); got != before {
		t.Fatalf("identical Redraw moved the pose: X = %v, want %v", got, before)
	}

	time.Sleep(5 * time.Millisecond)
	if err := p.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if after := motionTargetX(p, index); after <= before {
		t.Fatalf("the animation stopped after identical redraws: X = %v, want past %v", after, before)
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

func TestPageMotionIndexCaching(t *testing.T) {
	p := &Page{}

	// 1. Nil display or empty keyframes returns shared EmptyIndex with 0 allocs.
	emptyDisplay := &Display{
		Ops: []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
	}
	idx := updateMotionIndex(p, nil, emptyDisplay)
	if idx != motion.EmptyIndex() {
		t.Fatalf("expected emptyIndex, got %v", idx)
	}

	// 2. Display with keyframes builds a new index.
	keyframes := []layout.KeyframesRule{
		{
			Name: "fade",
			Keyframes: []layout.Keyframe{
				{Offset: 0, Declarations: []layout.KeyframeDeclaration{{Property: "opacity", Value: "0"}}},
				{Offset: 1, Declarations: []layout.KeyframeDeclaration{{Property: "opacity", Value: "1"}}},
			},
		},
	}
	d1 := &Display{
		Ops:       []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners:  []int{0},
		Boxes:     []layout.Box{{ID: "b1", OperationIDs: []uint64{1}}},
		Keyframes: keyframes,
	}

	idx1 := updateMotionIndex(p, nil, d1)
	if idx1 == nil || idx1 == motion.EmptyIndex() {
		t.Fatalf("expected non-empty index, got %v", idx1)
	}
	if p.motionIndex != idx1 {
		t.Fatalf("expected p.motionIndex to be cached, got %v", p.motionIndex)
	}

	// 3. Same keyframes (pointer equality) reuses the cached index.
	d2 := &Display{
		Ops:       []layout.DisplayOp{{ID: 1, Kind: layout.DisplayOpFillRect}},
		OpOwners:  []int{0},
		Boxes:     []layout.Box{{ID: "b1", OperationIDs: []uint64{1}}},
		Keyframes: keyframes, // identical slice
	}

	idx2 := updateMotionIndex(p, d1, d2)
	if idx2 != idx1 {
		t.Fatalf("expected reused index %p, got %p", idx1, idx2)
	}

	// 4. Zero allocs on repeated commits with same keyframes.
	allocs := testing.AllocsPerRun(100, func() {
		idx := updateMotionIndex(p, d1, d2)
		if idx == nil {
			t.Fatal("unexpected nil index")
		}
	})
	if allocs != 0 {
		t.Fatalf("allocs = %v, want 0", allocs)
	}

	// 5. Zero allocs when both have empty keyframes.
	p.motionIndex = nil
	allocsEmpty := testing.AllocsPerRun(100, func() {
		idx := updateMotionIndex(p, emptyDisplay, emptyDisplay)
		if idx != motion.EmptyIndex() {
			t.Fatal("expected emptyIndex")
		}
	})
	if allocsEmpty != 0 {
		t.Fatalf("allocsEmpty = %v, want 0", allocsEmpty)
	}
}

// TestTickAnimatedBoxesDoNotAllocatePerBox pins the steady-state tick cost of
// a CSS transform animation across many boxes. A tick used to blend and
// allocate a function list per animated box that only the 3D restamp pass
// reads, so 64 boxes cost 256 allocations.
func TestTickAnimatedBoxesDoNotAllocatePerBox(t *testing.T) {
	var cells strings.Builder
	for i := range 64 {
		fmt.Fprintf(&cells, `<div id="s%d" class="cell">cell %d</div>`, i, i)
	}
	p, err := New(Config{
		HTML: `<style>@keyframes slide{from{opacity:0.4;transform:translateX(0px)}to{opacity:1;transform:translateX(12px)}}` +
			`.cell{width:84px;height:22px;animation:slide 1s linear infinite}</style>` + cells.String(),
		Width: 760, Height: 520,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if err := p.Tick(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if got := p.Motion().Active; got != 64 {
		t.Fatalf("active players = %d, want 64", got)
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = p.Tick(ctx) }); allocs > 8 {
		t.Fatalf("a tick over 64 animated boxes allocated %v times, want a handful", allocs)
	}
}
