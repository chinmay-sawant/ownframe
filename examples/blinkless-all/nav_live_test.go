package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"
)

// newRig opens the nav tab with its live session and runs the first frame.
func newRig(t *testing.T, tab string) *liveRig {
	t.Helper()

	r, err := newLiveRig(context.Background(), newTestApp(t), tab, 1000, 800)
	if err != nil {
		t.Fatal(err)
	}

	must(t, r.step(0))

	return r
}

func must(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 0.6 }

func (r *liveRig) style(id string) string { return r.nav.written(id).Style }

func TestLiveFocusArrowKeys(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("n-a"))
	must(t, r.key("ArrowDown"))

	if r.nav.st.focus["focus"] != "n-f" {
		t.Fatalf("ArrowDown focus = %q, want n-f (nav-down: #n-f)", r.nav.st.focus["focus"])
	}

	if !strings.Contains(r.style("n-f"), "box-shadow") || strings.Contains(r.style("n-a"), "box-shadow") {
		t.Errorf("the focus ring must sit on n-f only: n-f=%q n-a=%q", r.style("n-f"), r.style("n-a"))
	}

	if got := r.textOf("chip-focus"); got != "moved n-a -> n-f via nav-down" {
		t.Errorf("chip = %q", got)
	}

	must(t, r.key("ArrowLeft"))

	if r.nav.st.focus["focus"] != "n-e" || !strings.Contains(r.textOf("chip-focus"), "via spatial") {
		t.Errorf("ArrowLeft: focus %q chip %q", r.nav.st.focus["focus"], r.textOf("chip-focus"))
	}

	// nav-left on n-b names #n-e with the current keyword.
	must(t, r.tap("n-b"))
	must(t, r.key("ArrowLeft"))

	if r.nav.st.focus["focus"] != "n-e" || !strings.Contains(r.textOf("chip-focus"), "via nav-left") {
		t.Errorf("nav-left: %q", r.textOf("chip-focus"))
	}
}

func TestLiveFocusOnScreenArrowsDoTheSame(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("n-a"))
	must(t, r.tap("arr-focus-down"))

	if r.nav.st.focus["focus"] != "n-f" || r.textOf("chip-focus") != "moved n-a -> n-f via nav-down" {
		t.Errorf("on-screen Down: %q %q", r.nav.st.focus["focus"], r.textOf("chip-focus"))
	}

	must(t, r.tap("arr-focus-left"))
	must(t, r.tap("arr-focus-up"))

	if r.nav.st.focus["focus"] != "n-b" {
		t.Errorf("Left then Up reached %q, want n-b", r.nav.st.focus["focus"])
	}
}

func TestLiveGridVersusNormalContainAndAction(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("gn-a"))
	must(t, r.tap("arr-grid-down"))

	if r.nav.st.focus["grid"] != "gn-c" {
		t.Errorf("normal function: gn-a down reached %q, want gn-c", r.nav.st.focus["grid"])
	}

	must(t, r.tap("gg-a"))
	must(t, r.tap("arr-grid-down"))

	if r.nav.st.focus["grid"] != "gg-a" || !strings.Contains(r.textOf("chip-grid"), "no target") {
		t.Errorf("grid function: %q %q", r.nav.st.focus["grid"], r.textOf("chip-grid"))
	}

	must(t, r.tap("c-b"))
	must(t, r.tap("arr-contain-down"))

	if r.nav.st.focus["contain"] != "c-b" || !strings.Contains(r.textOf("chip-contain"), "no target") {
		t.Errorf("contain: %q %q", r.nav.st.focus["contain"], r.textOf("chip-contain"))
	}
}

func TestLiveCaretBlinkTypingAndShapes(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("cr-bar"))
	must(t, r.typ("abc"))

	color := func(id string) string {
		s := r.style(id + "-c")

		return s[strings.LastIndex(s, "background:")+len("background:"):]
	}

	if color("cr-bar") != "#1f2937" || color("cr-under") != "#1f2937" || color("cr-block") != "#d00000" {
		t.Fatalf("at 0ms the carets are on: %q %q %q", color("cr-bar"), color("cr-under"), color("cr-block"))
	}

	if got := r.textOf("cr-bar-t"); got != "type hereabc" {
		t.Errorf("typed text = %q", got)
	}

	// The caret moved one char cell right per typed letter.
	if left := 4 + 12*liveCharW; !strings.Contains(r.style("cr-bar-c"), "left:"+trimFloat(left)+"px") {
		t.Errorf("caret position: %q", r.style("cr-bar-c"))
	}

	must(t, r.idle(500*time.Millisecond))

	if color("cr-bar") != "transparent" || color("cr-under") != "transparent" || color("cr-autoblock") != "transparent" {
		t.Errorf("at 500ms the automatic carets are off: %q", color("cr-bar"))
	}

	if color("cr-block") != "#d00000" {
		t.Error("the manual caret must stay on at 500ms")
	}

	must(t, r.idle(500*time.Millisecond))

	if color("cr-bar") != "#1f2937" || color("cr-block") != "#d00000" {
		t.Errorf("at 1000ms the carets are on again: %q", color("cr-bar"))
	}

	// Shapes: a bar is thin, a block is a cell, an underscore is a short line low.
	if !strings.Contains(r.style("cr-bar-c"), "width:2px;height:16px") ||
		!strings.Contains(r.style("cr-autoblock-c"), "width:8.4013671875px;height:16px") ||
		!strings.Contains(r.style("cr-under-c"), "width:8.4013671875px;height:2px") {
		t.Errorf("shapes: bar %q block %q under %q", r.style("cr-bar-c"), r.style("cr-autoblock-c"), r.style("cr-under-c"))
	}

	// Backspace removes the last letter and the caret steps back.
	must(t, r.key("Backspace"))

	if got := r.textOf("cr-bar-t"); got != "type herea"+"b" {
		t.Errorf("after Backspace %q", got)
	}
}

func trimFloat(v float64) string {
	return strings.TrimRight(strings.TrimRight(strings.Replace(formatFloat(v), ",", ".", 1), "0"), ".")
}

func formatFloat(v float64) string {
	return strings.TrimSpace(strings.Replace(sprintG(v), "e+", "e", 1))
}

func TestLiveFieldSizingGrowsOnlyTheContentField(t *testing.T) {
	r := newRig(t, "nav")

	contentBefore, fixedBefore := r.box("fs-content").W, r.box("fs-fixed").W

	must(t, r.tap("fs-content"))
	must(t, r.typ("xyz"))

	// Three typed letters and the caret, which takes a cell while the field has focus.
	if grown := r.box("fs-content").W - contentBefore; !near(grown, 4*liveCharW) {
		t.Errorf("content field grew by %g, want %g", grown, 4*liveCharW)
	}

	must(t, r.tap("fs-fixed"))
	must(t, r.typ("xyzxyz"))

	if got := r.box("fs-fixed").W; !near(got, fixedBefore) {
		t.Errorf("fixed field changed from %g to %g", fixedBefore, got)
	}

	must(t, r.tap("fs-content"))
	must(t, r.key("Backspace"))
	must(t, r.key("Backspace"))

	// One typed letter is left, plus the caret.
	if grown := r.box("fs-content").W - contentBefore; !near(grown, 2*liveCharW) {
		t.Errorf("after two Backspaces the content field is %g wider, want %g", grown, 2*liveCharW)
	}
}

func TestLiveInputSecurity(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("pw-auto"))
	must(t, r.typ("ab"))

	if got := r.textOf("pw-auto"); !strings.HasPrefix(got, strings.Repeat("•", 9)) || strings.Contains(got, "hunter") {
		t.Errorf("auto must show bullets, got %q", got)
	}

	must(t, r.tap("pw-none"))
	must(t, r.typ("ab"))

	if got := r.textOf("pw-none"); !strings.HasPrefix(got, "hunter2ab") {
		t.Errorf("none must show the text, got %q", got)
	}
}

func TestLiveUserSelectHighlight(t *testing.T) {
	r := newRig(t, "nav")

	// 'word' starts at char 7; aim at its middle.
	fx := (navTextLeft + 8.5*liveCharW) / 300

	must(t, r.press("us-text", fx, 0.5))
	must(t, r.press("us-all", fx, 0.5))
	must(t, r.press("us-none", fx, 0.5))

	hl := func(id string) string { return r.style(id + "-h") }

	if want := "left:" + trimFloat(navTextLeft+7*liveCharW) + "px;width:" + trimFloat(4*liveCharW) + "px"; !strings.Contains(hl("us-text"), want) {
		t.Errorf("word highlight %q, want %q", hl("us-text"), want)
	}

	if !strings.Contains(hl("us-all"), "left:4px;width:"+trimFloat(16*liveCharW)+"px") {
		t.Errorf("all highlight %q", hl("us-all"))
	}

	if !strings.Contains(hl("us-none"), "width:0px") {
		t.Errorf("none highlight %q", hl("us-none"))
	}
}

func TestLiveTextInflationToggle(t *testing.T) {
	r := newRig(t, "nav")

	for _, id := range []string{"ta-none", "ta-120", "ta-wk", "ta-auto"} {
		if !strings.Contains(r.style(id), "font-size:14px") {
			t.Errorf("%s starts at 14px: %q", id, r.style(id))
		}
	}

	must(t, r.tap("ta-btn"))

	for id, size := range map[string]string{"ta-none": "14", "ta-120": "16.8", "ta-wk": "21", "ta-auto": "21"} {
		if !strings.HasSuffix(r.style(id), "font-size:"+size+"px") {
			t.Errorf("inflated %s: %q, want %spx", id, r.style(id), size)
		}
	}

	must(t, r.tap("ta-btn"))

	if !strings.HasSuffix(r.style("ta-120"), "font-size:14px") {
		t.Errorf("toggled off: %q", r.style("ta-120"))
	}
}

func TestLiveSliderDirections(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.press("sl-h", 0.25, 0.5))

	if !near(r.nav.st.slider["sl-h"], 0.25) || !strings.Contains(r.style("sl-h-t"), "left:47.5px") {
		t.Errorf("left-to-right thumb: %v %q", r.nav.st.slider["sl-h"], r.style("sl-h-t"))
	}

	if !strings.Contains(r.textOf("chip-slider"), "sl-h 25") {
		t.Errorf("value %q", r.textOf("chip-slider"))
	}

	// Dragging keeps following the pointer while the button is down.
	must(t, r.dragTo("sl-h", 0.6, 0.5))

	if !strings.Contains(r.textOf("chip-slider"), "sl-h 60") {
		t.Errorf("drag value %q", r.textOf("chip-slider"))
	}

	must(t, r.release())
	must(t, r.dragTo("sl-h", 0.9, 0.5))

	if !strings.Contains(r.textOf("chip-slider"), "sl-h 60") {
		t.Errorf("after release the thumb must stay: %q", r.textOf("chip-slider"))
	}

	// right-to-left counts down from the left: a quarter along reads 75.
	must(t, r.press("sl-rtl", 0.25, 0.5))
	must(t, r.release())

	if !strings.Contains(r.textOf("chip-slider"), "sl-rtl 75") {
		t.Errorf("right-to-left %q", r.textOf("chip-slider"))
	}

	// bottom-to-top counts up from the bottom: a quarter down reads 75.
	must(t, r.press("sl-v", 0.5, 0.25))
	must(t, r.release())

	if !strings.Contains(r.textOf("chip-slider"), "sl-v 75") {
		t.Errorf("bottom-to-top %q", r.textOf("chip-slider"))
	}
}

func TestLiveWindowDragOnlyFromTheDraggableRegion(t *testing.T) {
	r := newRig(t, "nav")

	start := r.box("drag-win")

	must(t, r.press("drag-title", 0.5, 0.5))
	must(t, r.dragTo("drag-area", 0.7, 0.9))
	must(t, r.release())

	moved := r.box("drag-win")
	if moved.X-start.X < 50 || moved.Y-start.Y < 20 {
		t.Fatalf("window did not follow the pointer: %v -> %v", start, moved)
	}

	// Window-drag none on the close box and no value on the body: no drag.
	for _, id := range []string{"drag-btn", "drag-body"} {
		before := r.box("drag-win")

		must(t, r.press(id, 0.5, 0.5))
		must(t, r.dragTo("drag-area", 0.1, 0.1))
		must(t, r.release())

		after := r.box("drag-win")
		if !near(after.X, before.X) || !near(after.Y, before.Y) {
			t.Errorf("pressing %s moved the window from %v to %v", id, before, after)
		}

		if !strings.Contains(r.textOf("chip-drag"), "no drag") {
			t.Errorf("%s chip %q", id, r.textOf("chip-drag"))
		}
	}
}

func TestLiveWillChangeAnimationAndBadge(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("wc"))
	must(t, r.idle(300*time.Millisecond))

	if s := r.style("wc"); !strings.Contains(s, "transform:translateX(") || !strings.Contains(s, "opacity:0.") {
		t.Errorf("wc must animate: %q", s)
	}

	must(t, r.idle(time.Second))
	must(t, r.tap("wc2"))
	must(t, r.idle(time.Second+200*time.Millisecond))

	cached, plain := r.textOf("wc-badge"), r.textOf("wc2-badge")
	if !strings.Contains(cached, "kept on a cached layer") || !strings.Contains(cached, "painted 1 time") {
		t.Errorf("transform hint should paint once: %q", cached)
	}

	if !strings.Contains(plain, "no cached layer") || !strings.Contains(plain, "(every frame)") {
		t.Errorf("margin hint has no layer and repaints every frame: %q", plain)
	}

	// Same frames moved, very different paint cost: that is what the hint buys.
	var framesCached, framesPlain, paintsPlain int

	_, cachedTail, _ := strings.Cut(cached, "Moved over")
	_, plainTail, _ := strings.Cut(plain, "Moved over")

	if _, err := fmt.Sscanf(cachedTail, "%d frames", &framesCached); err != nil {
		t.Fatal(err)
	}

	if _, err := fmt.Sscanf(plainTail, "%d frames, painted %d times", &framesPlain, &paintsPlain); err != nil {
		t.Fatal(err)
	}

	if framesCached < 5 || paintsPlain != framesPlain {
		t.Errorf("cached moved %d frames; plain moved %d and painted %d", framesCached, framesPlain, paintsPlain)
	}

	if s := r.style("wc2"); strings.Contains(s, "transform") {
		t.Errorf("the animation must end: %q", s)
	}
}

func TestLiveAppearancePressedLooks(t *testing.T) {
	r := newRig(t, "nav")

	if !strings.Contains(r.style("ap-none"), "background:transparent") || !strings.Contains(r.style("ap-btn"), "border:2px solid #9ca3af") {
		t.Fatalf("rest looks: none %q button %q", r.style("ap-none"), r.style("ap-btn"))
	}

	must(t, r.press("ap-btn", 0.5, 0.5))

	if !strings.Contains(r.style("ap-btn"), "background:#cbd5e1") || !strings.Contains(r.textOf("chip-appear"), "ap-btn button pressed") {
		t.Errorf("pressed button %q", r.style("ap-btn"))
	}

	must(t, r.release())
	must(t, r.press("ap-none", 0.5, 0.5))

	if !strings.Contains(r.style("ap-none"), "background:#e5e7eb") || strings.Contains(r.style("ap-none"), "border:2px") {
		t.Errorf("pressed plain %q", r.style("ap-none"))
	}

	must(t, r.release())

	if !strings.Contains(r.style("ap-none"), "background:transparent") {
		t.Errorf("released plain %q", r.style("ap-none"))
	}
}

func TestLiveImageAnimationFrames(t *testing.T) {
	r := newRig(t, "nav")

	frame := func(id string) string { return r.textOf(id) }

	// The page ticked at 0ms: frame 1 (index 0) of the normal image.
	if frame("ia-normal") != "1" || frame("ia-stopped") != "1" || frame("ia-paused") != "3" {
		t.Fatalf("at 0ms: %s %s %s", frame("ia-normal"), frame("ia-stopped"), frame("ia-paused"))
	}

	for _, c := range []struct {
		advance time.Duration
		want    string
	}{{200, "2"}, {200, "3"}, {900, "3"}} { // 200, 400, 1300ms
		must(t, r.step(c.advance*time.Millisecond))

		if frame("ia-normal") != c.want {
			t.Errorf("normal frame at +%v = %s, want %s", c.advance, frame("ia-normal"), c.want)
		}

		if frame("ia-stopped") != "1" || frame("ia-paused") != "3" {
			t.Errorf("stopped %s paused %s must not move", frame("ia-stopped"), frame("ia-paused"))
		}
	}

	// Pause freezes the normal image; resume continues from the same frame.
	must(t, r.tap("ia-btn"))

	frozen := frame("ia-normal")

	must(t, r.idle(600*time.Millisecond))

	if frame("ia-normal") != frozen || r.textOf("ia-btn") != "Resume normal" {
		t.Errorf("paused image moved: %s -> %s", frozen, frame("ia-normal"))
	}

	must(t, r.tap("ia-btn"))

	if frame("ia-normal") != frozen {
		t.Errorf("resume must continue from frame %s, got %s", frozen, frame("ia-normal"))
	}

	must(t, r.idle(400*time.Millisecond))

	if frame("ia-normal") == frozen {
		t.Error("after resume the frames must cycle again")
	}
}

func TestLiveSessionIdleOnOtherTabAndResetsOnReturn(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("n-a"))
	must(t, r.key("ArrowDown"))
	must(t, r.tap("fs-content"))
	must(t, r.typ("zzz"))

	if r.nav.patches == 0 || !r.nav.active {
		t.Fatal("the session must be live on the nav tab")
	}

	app := newTestApp(t)

	must(t, r.page.Load(r.ctx, app.Page("read")))

	before := r.nav.patches

	must(t, r.press("r-flex", 0.5, 0.5))
	must(t, r.key("ArrowDown"))
	must(t, r.typ("qqq"))
	must(t, r.idle(time.Second))

	if r.nav.active || r.nav.patches != before {
		t.Errorf("on another tab the session must do nothing: active=%v patches %d -> %d", r.nav.active, before, r.nav.patches)
	}

	// Back on the nav tab the state starts over.
	must(t, r.page.Load(r.ctx, app.Page("nav")))
	must(t, r.step(0))

	if !r.nav.active || r.nav.st.focus["focus"] != "n-a" || r.nav.st.texts["fs-content"] != "hello" {
		t.Errorf("state must reset on return: %v %v", r.nav.st.focus, r.nav.st.texts["fs-content"])
	}
}

func TestLiveSessionStartsOverWhenTheTabIsReloaded(t *testing.T) {
	r := newRig(t, "nav")

	must(t, r.tap("n-a"))
	must(t, r.key("ArrowDown"))

	if r.nav.st.focus["focus"] != "n-f" {
		t.Fatal("setup")
	}

	// Pressing the open tab's button loads the same document again.
	must(t, r.page.Load(r.ctx, newTestApp(t).Page("nav")+"<!-- reloaded -->"))
	must(t, r.step(0))

	if r.nav.st.focus["focus"] != "n-a" {
		t.Errorf("a reloaded page must reset the session, focus %q", r.nav.st.focus["focus"])
	}
}

func sprintG(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

// TestLiveActionScrollRoundTrip is the user's check: Down and Right move the
// sheet 40px each way, Up and Left bring it back, and the ends hold.
func TestLiveActionScrollRoundTrip(t *testing.T) {
	r := newRig(t, "nav")
	origin := r.box("s-a")

	at := func(want [2]float64) {
		t.Helper()

		if r.nav.st.scroll != want || r.nav.st.focus["action"] == "s-b" {
			t.Fatalf("offset %v, want %v (%s)", r.nav.st.scroll, want, r.textOf("chip-action"))
		}

		got := r.box("s-a")
		if !near(origin.X-got.X, want[0]) || !near(origin.Y-got.Y, want[1]) {
			t.Fatalf("s-a moved (%g,%g) on screen, want (%g,%g)", origin.X-got.X, origin.Y-got.Y, want[0], want[1])
		}
	}

	must(t, r.tap("arr-action-down"))
	at([2]float64{0, 40})
	must(t, r.tap("arr-action-right"))
	at([2]float64{40, 40})
	must(t, r.tap("arr-action-up"))
	at([2]float64{40, 0})
	must(t, r.tap("arr-action-left"))
	at([2]float64{0, 0})

	// At the start Up and Left change nothing and say so.
	must(t, r.tap("arr-action-up"))
	must(t, r.tap("arr-action-left"))
	at([2]float64{0, 0})

	if !strings.Contains(r.textOf("chip-action"), "already at the end") {
		t.Errorf("chip at the edge: %q", r.textOf("chip-action"))
	}

	// The keyboard does the same, and the far ends clamp to the sheet.
	for i := 0; i < 12; i++ {
		must(t, r.key("ArrowDown"))
		must(t, r.key("ArrowRight"))
	}

	at([2]float64{navSheetW - navViewW, navSheetH - navViewH})
	must(t, r.key("ArrowUp"))
	must(t, r.key("ArrowLeft"))
	at([2]float64{navSheetW - navViewW - 40, navSheetH - navViewH - 40})
}

// TestLiveFieldsStayOnOneLine checks a typed field never wraps: the content
// field grows to the right, and the fixed and password fields keep their width
// and scroll their text so the caret stays at the right edge.
func TestLiveFieldsStayOnOneLine(t *testing.T) {
	r := newRig(t, "nav")

	oneLine := func(id string) {
		t.Helper()

		got := r.box(id)
		if got.H > 30 {
			t.Fatalf("%s is %gpx tall: its text wrapped", id, got.H)
		}
	}

	must(t, r.tap("fs-content"))
	oneLine("fs-content")

	wasContent := r.box("fs-content").W

	must(t, r.typ("abcdefghijklmn"))
	oneLine("fs-content")

	if r.box("fs-content").W <= wasContent {
		t.Errorf("content field did not grow to the right: %g then %g", wasContent, r.box("fs-content").W)
	}

	must(t, r.tap("fs-fixed"))
	must(t, r.typ("abcdefghijklmnopqrstuvwxyz"))
	oneLine("fs-fixed")

	if w := r.box("fs-fixed").W; !near(w, 150) {
		t.Errorf("fixed field changed width to %g", w)
	}

	if shown := r.textOf("fs-fixed"); !strings.HasSuffix(shown, "|") || len([]rune(shown)) > 16 {
		t.Errorf("fixed field should show the end of its text with the caret last: %q", shown)
	}

	must(t, r.tap("pw-auto"))
	must(t, r.typ("abcdefghijklmnopqrstuvwxyz"))
	oneLine("pw-auto")

	if shown := r.textOf("pw-auto"); !strings.HasSuffix(shown, "|") || len([]rune(shown)) > 16 {
		t.Errorf("password field should scroll with the caret last: %q", shown)
	}
}

// TestLiveLongTextNeverLeavesItsBox types 30 characters into the fixed and
// password fields. With or without focus, the text drawn must fit inside the
// box, and a focused field must end with the caret.
func TestLiveLongTextNeverLeavesItsBox(t *testing.T) {
	r := newRig(t, "nav")
	long := strings.Repeat("Wm1iX", 6)

	fits := func(id string) {
		t.Helper()

		box := r.box(id)
		shown := r.textOf(id)

		if got := float64(len([]rune(shown))) * liveCharW; got > box.W-2 {
			t.Errorf("%s draws %q, %gpx of text in a %gpx box", id, shown, got, box.W)
		}
	}

	for _, id := range []string{"fs-fixed", "pw-auto", "pw-none"} {
		must(t, r.tap(id))
		must(t, r.typ(long))
		fits(id)

		if !strings.HasSuffix(r.textOf(id), "|") {
			t.Errorf("%s has focus, so the caret should be last: %q", id, r.textOf(id))
		}
	}

	// Focus has moved on from the first two, which now show their start.
	for _, id := range []string{"fs-fixed", "pw-auto"} {
		fits(id)

		if strings.HasSuffix(r.textOf(id), "|") {
			t.Errorf("%s lost focus but still shows a caret: %q", id, r.textOf(id))
		}
	}

	if shown := r.textOf("fs-fixed"); !strings.HasPrefix(shown, "helloWm1") {
		t.Errorf("an unfocused fixed field shows the start of its text: %q", shown)
	}

	// The 30 characters were all taken; nothing was thrown away on the way in.
	if got := r.nav.st.texts["fs-fixed"]; got != "hello"+long {
		t.Errorf("fixed field holds %q, want hello + the 30 typed characters", got)
	}
}
