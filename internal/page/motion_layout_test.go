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

	indices := cssmotion.OperationIndices(p.Display(), p.boxByID(id))
	if len(indices) == 0 {
		t.Fatalf("%s owns no operation", id)
	}

	return indices[0]
}

// textOpOf returns the first text operation the element with id owns, or nil.
func textOpOf(p *Page, id string) *DisplayOp {
	d := p.Display()
	for _, index := range cssmotion.OperationIndices(d, p.boxByID(id)) {
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
	before := p.boxByID("grow")
	nextBefore := p.boxByID("next")

	tickLayout(t, p)

	after := p.boxByID("grow")
	nextAfter := p.boxByID("next")
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
	if again := p.boxByID("grow").W; again < after.W {
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

	aBefore, bBefore := p.boxByID("a").W, p.boxByID("b").W

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

	aFirst, bFirst := p.boxByID("a").W, p.boxByID("b").W
	switch {
	case aFirst == aBefore && bFirst == bBefore:
		t.Fatal("neither element moved on the first frame")
	case aFirst != aBefore && bFirst != bBefore:
		t.Fatal("both elements relaid out in one frame, so the budget was exceeded")
	}

	tickLayout(t, p)
	if aFirst == aBefore {
		if got := p.boxByID("a").W; got == aFirst {
			t.Fatalf("the queued first element never ran: width %v", got)
		}
	}
	if bFirst == bBefore {
		if got := p.boxByID("b").W; got == bFirst {
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
	if got := p.boxByID("grow").W; got != 40 {
		t.Fatalf("grow width = %v on the redraw frame, want the base 40", got)
	}
	if p.restampRedraw {
		t.Fatal("the owed retained-state redraw stayed queued")
	}
	if p.Generation() == gen {
		t.Fatal("the first tick did not redraw with retained layout state")
	}

	tickLayout(t, p)
	if got := p.boxByID("grow").W; got <= 40 {
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

	next := p.boxByID("next")
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
	if got := p.boxByID("b").W; got != 40 {
		t.Fatalf("b width = %v on the frame it waited, want the base 40", got)
	}

	time.Sleep(20 * time.Millisecond)
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.boxByID("b").W; got <= 40 {
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

			err = p.applyLayoutRequest(context.Background(), motion.LayoutRequest{
				ElementID: "target",
				Overrides: []motion.LayoutOverride{{Property: test.property, Value: test.value}},
			})
			if err != nil {
				t.Fatalf("%s: %v: %s", test.property, err, test.value)
			}
		})
	}
}
