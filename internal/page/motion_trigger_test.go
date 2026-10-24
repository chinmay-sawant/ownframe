package page

import (
	"context"
	"math"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

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

	box := p.boxByID(id)
	if box.ID == "" {
		t.Fatalf("box %q is missing", id)
	}
	indices := cssmotion.OperationIndices(p.Display(), box)
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

	box := p.boxByID(id)
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
	if got := p.formState().FocusID; got != "play" {
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
