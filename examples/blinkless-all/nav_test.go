package main

import (
	"context"
	"strings"
	"testing"
)

func demoModel(t *testing.T, markup string) navModel {
	t.Helper()

	m, err := navLayout(context.Background(), `<!DOCTYPE html><html><body style="margin:0">`+markup+`</body></html>`, 600)
	if err != nil {
		t.Fatal(err)
	}

	return m
}

// moveID returns the id a key press reaches, "" for none.
func moveID(m navModel, from string, dir navDir) (string, navResult) {
	r := m.Move(m.ByID(from), dir, navFocusable)
	if r.Target < 0 {
		return "", r
	}

	return m.Items[r.Target].ID, r
}

func TestNavExplicitTargetAndSpatialFallback(t *testing.T) {
	m := demoModel(t, navDemoHTML)

	// nav-down: #n-f jumps over n-d, the nearest button below n-a.
	if id, r := moveID(m, "n-a", navDown); id != "n-f" || r.Via != "nav-down" {
		t.Errorf("n-a down = %q via %q, want n-f via nav-down", id, r.Via)
	}

	// No nav-right on n-a: the nearest button on the right.
	if id, r := moveID(m, "n-a", navRight); id != "n-b" || r.Via != "spatial" {
		t.Errorf("n-a right = %q via %q, want n-b via spatial", id, r.Via)
	}

	// nav-left with the current frame keyword.
	if id, _ := moveID(m, "n-b", navLeft); id != "n-e" {
		t.Errorf("n-b left = %q, want n-e", id)
	}

	if id, _ := moveID(m, "n-e", navUp); id != "n-b" {
		t.Errorf("n-e up = %q, want n-b", id)
	}

	// nav-up and nav-right through the same path, plus an unknown id and a
	// foreign frame name, which fall back to the spatial search.
	m2 := demoModel(t, `<div style="position:relative;height:90px">
	<button id="n-a" style="nav-up:#n-c;nav-right:#nowhere;nav-left:#n-b 'other';position:absolute;left:100px;top:40px;width:40px;height:20px">a</button>
	<button id="n-b" style="position:absolute;left:20px;top:40px;width:40px;height:20px">b</button>
	<button id="n-c" style="position:absolute;left:200px;top:0;width:40px;height:20px">c</button>
	<button id="n-d" style="nav-right:#n-c root;position:absolute;left:300px;top:40px;width:40px;height:20px">d</button></div>`)

	if id, r := moveID(m2, "n-a", navUp); id != "n-c" || r.Via != "nav-up" {
		t.Errorf("nav-up = %q via %q", id, r.Via)
	}

	if id, r := moveID(m2, "n-a", navRight); id != "n-c" || r.Via != "spatial" {
		t.Errorf("unknown id must fall back to spatial, got %q via %q", id, r.Via)
	}

	if id, r := moveID(m2, "n-a", navLeft); id != "n-b" || r.Via != "spatial" {
		t.Errorf("foreign frame must fall back to spatial, got %q via %q", id, r.Via)
	}

	if id, r := moveID(m2, "n-d", navRight); id != "n-c" || r.Via != "nav-right" {
		t.Errorf("nav-right root = %q via %q", id, r.Via)
	}

	// auto means spatial.
	if id, r := moveID(demoModel(t, strings.Replace(navDemoHTML, "nav-down:#n-f", "nav-down:auto", 1)), "n-a", navDown); id != "n-d" || r.Via != "spatial" {
		t.Errorf("nav-down auto = %q via %q, want n-d", id, r.Via)
	}
}

func TestNavFunctionGridVersusNormal(t *testing.T) {
	m := demoModel(t, navDemoHTML)

	if id, _ := moveID(m, "gn-a", navDown); id != "gn-c" {
		t.Errorf("normal: gn-a down = %q, want gn-c", id)
	}

	if id, _ := moveID(m, "gg-a", navDown); id != "" {
		t.Errorf("grid: gg-a down = %q, want nothing (gg-c is not in the same column)", id)
	}
}

func TestNavContainStopsAtTheBox(t *testing.T) {
	m := demoModel(t, navDemoHTML)

	if id, _ := moveID(m, "c-a", navRight); id != "c-b" {
		t.Errorf("c-a right = %q, want c-b", id)
	}

	if id, _ := moveID(m, "c-b", navDown); id != "" {
		t.Errorf("contain: c-b down = %q, want nothing; o-x is outside the box", id)
	}

	free := demoModel(t, strings.Replace(navDemoHTML, `id="cbox" style="spatial-navigation-contain:contain;`, `id="cbox" style="`, 1))
	if id, _ := moveID(free, "c-b", navDown); id != "o-x" {
		t.Errorf("without contain c-b down = %q, want o-x", id)
	}
}

func TestNavActionScrollTurnsKeysIntoScroll(t *testing.T) {
	m := demoModel(t, navDemoHTML)

	_, r := moveID(m, "s-a", navDown)
	if r.Action != "scroll" || r.ScrollDY != scrollStep || r.Target != -1 {
		t.Errorf("scroll action: %+v", r)
	}

	plain := demoModel(t, strings.Replace(navDemoHTML, "spatial-navigation-action:scroll;", "", 1))
	if _, r = moveID(plain, "s-a", navRight); r.Action != "focus" || r.ScrollDX != 0 {
		t.Errorf("without the property the key moves focus: %+v", r)
	}
}

func TestNavCaret(t *testing.T) {
	m := demoModel(t, navDemoHTML)

	bar := navCaretOf(m.Items[m.ByID("cr-bar")].Declared)
	block := navCaretOf(m.Items[m.ByID("cr-block")].Declared)
	under := navCaretOf(m.Items[m.ByID("cr-under")].Declared)

	if _, _, w, h := bar.Rect(10, 20, 9, 16); w != caretBarWidth || h != 16 {
		t.Errorf("bar rect w=%g h=%g", w, h)
	}

	if x, y, w, h := block.Rect(10, 20, 9, 16); x != 10 || y != 20 || w != 9 || h != 16 {
		t.Errorf("block rect %g %g %g %g", x, y, w, h)
	}

	if _, y, w, h := under.Rect(10, 20, 9, 16); y != 34 || w != 9 || h != caretBarWidth {
		t.Errorf("underscore rect y=%g w=%g h=%g", y, w, h)
	}

	if block.Color != "#d00000" || !block.Manual || bar.Manual {
		t.Errorf("shorthand parts: %+v %+v", block, bar)
	}

	// caret-animation: auto blinks; manual stays on.
	if !under.Visible(0) || under.Visible(500) || !under.Visible(1000) {
		t.Error("automatic caret must blink every 500ms")
	}

	if !block.Visible(500) {
		t.Error("manual caret must not blink")
	}

	// A longhand beats the shorthand.
	c := navCaretOf(map[string]string{"caret": "block manual", "caret-shape": "underscore", "caret-animation": "auto"})
	if c.Shape != "underscore" || c.Manual {
		t.Errorf("longhand must win: %+v", c)
	}
}

func TestNavFieldSecuritySelectAndSize(t *testing.T) {
	m := demoModel(t, navDemoHTML)
	decl := func(id, p string) string { return m.Items[m.ByID(id)].Declared[p] }

	grown := navFieldWidth(decl("fs-content", "field-sizing"), "hello world", 8, 12, 40, 150)
	fixed := navFieldWidth(decl("fs-fixed", "field-sizing"), "hello world", 8, 12, 40, 150)

	if grown != 100 || fixed != 150 || navFieldWidth("content", "", 8, 12, 40, 150) != 40 {
		t.Errorf("field widths %g %g", grown, fixed)
	}

	if got := navShownText("password", decl("pw-auto", "input-security"), "hunter2"); got != "•••••••" {
		t.Errorf("auto must hide, got %q", got)
	}

	if got := navShownText("password", decl("pw-none", "input-security"), "hunter2"); got != "hunter2" {
		t.Errorf("none must show, got %q", got)
	}

	if got := navShownText("text", "auto", "abc"); got != "abc" {
		t.Errorf("a text field is never hidden, got %q", got)
	}

	const text = "pick a word here"
	for id, want := range map[string][2]int{"us-text": {7, 11}, "us-none": {7, 7}, "us-all": {0, len(text)}} {
		lo, hi := navSelect(decl(id, "-webkit-user-select"), text, 7)
		if lo != want[0] || hi != want[1] {
			t.Errorf("%s selects [%d,%d), want %v", id, lo, hi, want)
		}
	}

	scaled := map[string]float64{"ta-none": 14, "ta-120": 16.8, "ta-wk": 21}
	for id, want := range scaled {
		got := navScaleText(navTextAdjustOf(m.Items[m.ByID(id)].Declared), 14, 1.1)
		if d := got - want; d > 0.001 || d < -0.001 {
			t.Errorf("%s: 14px -> %g, want %g", id, got, want)
		}
	}

	if got := navScaleText("auto", 10, 1.5); got != 15 {
		t.Errorf("auto uses the host factor, got %g", got)
	}

	if navTextAdjustOf(map[string]string{"text-size-adjust": "none", "-webkit-text-size-adjust": "150%"}) != "none" {
		t.Error("the standard property must win")
	}
}

func TestNavSliderDragLayerAppearanceImage(t *testing.T) {
	m := demoModel(t, navDemoHTML)
	item := func(id string) navItem { return m.Items[m.ByID(id)] }

	h, r, v := item("sl-h"), item("sl-rtl"), item("sl-v")

	if got := navSliderValue(h.Declared["slider-orientation"], h.X, h.Y, h.W, h.H, h.X+h.W*0.25, h.Y, 0, 100); got != 25 {
		t.Errorf("left-to-right = %g", got)
	}

	if got := navSliderValue(r.Declared["slider-orientation"], r.X, r.Y, r.W, r.H, r.X+r.W*0.25, r.Y, 0, 100); got != 75 {
		t.Errorf("right-to-left = %g", got)
	}

	if got := navSliderValue(v.Declared["slider-orientation"], v.X, v.Y, v.W, v.H, v.X, v.Y+v.H*0.25, 0, 100); got != 75 {
		t.Errorf("bottom-to-top = %g", got)
	}

	if got := navSliderValue("auto", 0, 0, 10, 100, 0, 100, 0, 1); got != 0 {
		t.Errorf("auto on a tall track runs bottom to top, got %g", got)
	}

	if !m.navDragRegion(m.ByID("drag-title")) || m.navDragRegion(m.ByID("drag-btn")) || m.navDragRegion(m.ByID("pad")) {
		t.Error("window-drag: title drags, close button and plain boxes do not")
	}

	props, promote := navLayerHints(item("wc").Declared["will-change"])
	if !promote || len(props) != 2 || props[0] != "transform" {
		t.Errorf("will-change hints %v promote=%v", props, promote)
	}

	if _, p := navLayerHints("color, margin"); p {
		t.Error("color and margin do not need a layer")
	}

	if _, p := navLayerHints("auto"); p {
		t.Error("auto means no layer")
	}

	if navControlKind("button", "", item("ap-none").Declared["-webkit-appearance"]) != "plain" ||
		navControlKind("div", "", item("ap-btn").Declared["-webkit-appearance"]) != "button" ||
		navControlKind("button", "", "") != "button" {
		t.Error("appearance decides the control kind")
	}

	for id, want := range map[string]int{"ia-paused": 2, "ia-stopped": 0, "ia-normal": 2} {
		// 1300ms / 200ms = 6, mod 4 = 2; paused freezes at 450ms = frame 2.
		got := navImageFrame(item(id).Declared["image-animation"], 1300, 450, 200, 4)
		if got != want {
			t.Errorf("%s frame %d, want %d", id, got, want)
		}
	}

	if navImageFrame("paused", 1300, 650, 200, 4) != 3 {
		t.Error("paused must follow the pause time, not the clock")
	}
}

func TestNavDemosShowTheDecisions(t *testing.T) {
	for want, body := range map[string]string{
		"n-a down -&gt; n-f (nav-down)": navDemoFocus(),
		"gg-a down -&gt; nothing":       navDemoGrid(),
		"s-a down: scroll":              navDemoAction(),
		"\u2022\u2022\u2022":            navDemoSecurity(),
		"transform + opacity":           navDemoLayer(),
		"ia-paused: frame 3":            navDemoImage(),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("demo lacks %q", want)
		}
	}
}
