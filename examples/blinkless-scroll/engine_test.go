package main

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func near(t *testing.T, got, want float64) {
	t.Helper()

	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScrollProgress(t *testing.T) {
	v := Viewport{ScrollY: 500, H: 500, ContentH: 1500, W: 400, ContentW: 400}
	near(t, ScrollProgress(v, AxisY), 0.5)
	near(t, ScrollProgress(v, AxisX), 0) // cannot scroll on x
	v.ScrollY = 5000
	near(t, ScrollProgress(v, AxisY), 1)
}

func TestViewTimelineRanges(t *testing.T) {
	// viewport 400 tall, subject 100 tall at y=1000.
	subject := Rect{Y: 1000, W: 50, H: 100}
	at := func(scroll, ins0, ins1 float64) ViewState {
		return ViewTimeline(Viewport{ScrollY: scroll, H: 400, ContentH: 3000}, subject, AxisY, ins0, ins1)
	}
	// Subject top touches the bottom of the viewport at scroll 600.
	near(t, at(600, 0, 0).Progress("cover"), 0)
	near(t, at(600, 0, 0).Progress("entry"), 0)
	// Fully in at scroll 700.
	near(t, at(700, 0, 0).Progress("entry"), 1)
	near(t, at(650, 0, 0).Progress("entry"), 0.5)
	// Leaves the top at scroll 1100.
	near(t, at(1100, 0, 0).Progress("cover"), 1)
	near(t, at(1000, 0, 0).Progress("exit"), 0) // subject top at viewport top
	near(t, at(1100, 0, 0).Progress("exit"), 1)
	// Contain: subject smaller than port, so [100, 400] of a 500 trip.
	near(t, at(700, 0, 0).Progress("contain"), 0)
	near(t, at(1000, 0, 0).Progress("contain"), 1)
	// Insets shrink the scrollport: a 100px bottom inset delays entry by 100.
	near(t, at(700, 0, 100).Progress("entry"), 0)
	near(t, at(800, 0, 100).Progress("entry"), 1)
	// A 100px top inset makes the exit finish 100 later.
	near(t, at(900, 100, 0).Progress("cover"), 0.75)
}

func nodesWithScope() []Node {
	return []Node{
		{ID: "root", Parent: -1, Scope: Scope{Kind: ScopeNames, Names: []string{"--a"}}},                             // 0
		{ID: "decl", Parent: 0, Timelines: []Timeline{{Name: "--a", Kind: KindView}, {Name: "--b", Kind: KindView}}}, // 1
		{ID: "reader", Parent: 0}, // 2
		{ID: "outside", Parent: -1},
	}
}

func TestResolveScope(t *testing.T) {
	n := nodesWithScope()

	if _, at, ok := ResolveTimeline(n, 1, "--b"); !ok || at != 1 {
		t.Fatal("declarer sees its own name")
	}

	if _, at, ok := ResolveTimeline(n, 2, "--a"); !ok || at != 1 {
		t.Fatalf("scoped name is visible to a sibling: %v %v", at, ok)
	}

	if _, _, ok := ResolveTimeline(n, 2, "--b"); ok {
		t.Fatal("unscoped name must stay hidden from a sibling")
	}

	if _, _, ok := ResolveTimeline(n, 3, "--a"); ok {
		t.Fatal("a node outside the scope root must not see it")
	}

	n[0].Scope = Scope{Kind: ScopeAll}
	if _, _, ok := ResolveTimeline(n, 2, "--b"); !ok {
		t.Fatal("scope all hoists every name")
	}

	n[0].Scope = Scope{}
	if _, _, ok := ResolveTimeline(n, 2, "--a"); ok {
		t.Fatal("scope none hides it")
	}
}

func TestPointerProgress(t *testing.T) {
	box := Rect{X: 100, Y: 50, W: 200, H: 100}
	p, in := PointerProgress(150, 60, box, AxisX)
	near(t, p, 0.25)

	if !in {
		t.Fatal("inside")
	}

	p, _ = PointerProgress(150, 100, box, AxisY)
	near(t, p, 0.5)

	p, in = PointerProgress(900, 0, box, AxisX)
	near(t, p, 1)

	if in {
		t.Fatal("outside")
	}
}

func TestTimelineTriggerHysteresis(t *testing.T) {
	space := RangeSpace{Total: 100, Named: map[string][2]float64{"cover": {0, 100}}}
	tr := TimelineTrigger{
		ActivationStart: RangeBound{HasPercent: true, Percent: 20},
		ActivationEnd:   RangeBound{HasPercent: true, Percent: 80},
		ActiveStart:     RangeBound{HasPercent: true, Percent: 10},
		ActiveEnd:       RangeBound{HasPercent: true, Percent: 90},
	}

	var st TriggerState

	steps := []struct {
		pos  float64
		want bool
	}{
		{5, false}, {15, false}, // inside active but never activated
		{25, true},  // enters activation
		{85, true},  // left activation, still active
		{95, false}, // left active range
		{85, false}, // activation ends at 80, so no re-arm
		{50, true},
		{5, false},
	}
	for _, s := range steps {
		if got := tr.Evaluate(&st, s.pos, space); got != s.want {
			t.Fatalf("pos %v got %v want %v", s.pos, got, s.want)
		}
	}
}

func TestTriggerNamedRange(t *testing.T) {
	st := ViewTimeline(Viewport{ScrollY: 650, H: 400, ContentH: 3000}, Rect{Y: 1000, H: 100}, AxisY, 0, 0)
	tr := TimelineTrigger{
		ActivationStart: RangeBound{Name: "entry", HasPercent: true, Percent: 50},
		ActivationEnd:   RangeBound{Name: "exit"},
		ActiveStart:     RangeBound{Auto: true},
		ActiveEnd:       RangeBound{Auto: true},
	}

	var s TriggerState
	if !tr.Evaluate(&s, st.T, st.Space) {
		t.Fatal("T=50 is the entry midpoint, so it activates")
	}
}

func TestEventTriggers(t *testing.T) {
	e := NewEventTriggers([]EventTrigger{
		{Name: "--press", Start: []string{"click"}, End: []string{"keydown"}},
		{Name: "--flip", Start: []string{"click"}, End: []string{"click"}},
	})

	if got := e.Dispatch("click"); len(got) != 2 || !e.Active("--press") || !e.Active("--flip") {
		t.Fatalf("click starts both: %v", got)
	}

	e.Dispatch("keydown")

	if e.Active("--press") || !e.Active("--flip") {
		t.Fatal("keydown ends press only")
	}

	e.Dispatch("click") // press on, flip toggles off

	if !e.Active("--press") || e.Active("--flip") {
		t.Fatal("click: press on, flip toggled off")
	}

	if len(e.Dispatch("hover")) != 0 {
		t.Fatal("unknown event changes nothing")
	}
}

type fakeClock struct{ t time.Time }

func (c *fakeClock) Now() time.Time { return c.t }

func TestInterestDelay(t *testing.T) {
	c := &fakeClock{t: time.Unix(0, 0)}
	in := NewInterest(InterestDelay{StartMs: 200, EndMs: 500}, c)

	in.Hover(true)

	if in.Tick() {
		t.Fatal("not yet")
	}

	c.t = c.t.Add(199 * time.Millisecond)

	if in.Tick() {
		t.Fatal("199ms is early")
	}

	c.t = c.t.Add(1 * time.Millisecond)

	if !in.Tick() {
		t.Fatal("200ms starts interest")
	}

	in.Hover(false)
	c.t = c.t.Add(400 * time.Millisecond)

	if !in.Tick() {
		t.Fatal("end delay is 500ms")
	}

	// Coming back before the end timer runs out cancels it.
	in.Hover(true)
	c.t = c.t.Add(time.Second)

	if !in.Tick() {
		t.Fatal("re-hover keeps interest")
	}

	in.Hover(false)
	c.t = c.t.Add(500 * time.Millisecond)

	if in.Tick() {
		t.Fatal("500ms ends interest")
	}

	// A hover that ends before the start delay never turns on.
	in.Hover(true)
	c.t = c.t.Add(100 * time.Millisecond)
	in.Hover(false)
	c.t = c.t.Add(time.Second)

	if in.Tick() {
		t.Fatal("short hover must not trigger")
	}
}

func TestInterestNormal(t *testing.T) {
	c := &fakeClock{t: time.Unix(0, 0)}
	in := NewInterest(InterestDelay{StartNormal: true, EndNormal: true}, c)
	in.Hover(true)
	c.t = c.t.Add(NormalStartDelay)

	if !in.Tick() {
		t.Fatal("normal start delay")
	}
}

func inertNodes() []Node {
	return []Node{
		{ID: "page", Parent: -1, Rect: Rect{W: 500, H: 500}},
		{ID: "inert", Parent: 0, Rect: Rect{W: 100, H: 100}, Inert: true, Focusable: true},
		{ID: "child", Parent: 1, Rect: Rect{W: 50, H: 50}, Focusable: true}, // inherits inert
		{ID: "live", Parent: 0, Rect: Rect{X: 200, W: 100, H: 100}, Focusable: true},
	}
}

func TestInertHitAndFocus(t *testing.T) {
	n := inertNodes()

	if got := HitTest(n, 10, 10); got != 0 {
		t.Fatalf("click on inert subtree falls to the page, got %d", got)
	}

	if got := HitTest(n, 210, 10); got != 3 {
		t.Fatalf("live hit got %d", got)
	}

	if got := NextFocus(n, -1, 1); got != 3 {
		t.Fatalf("focus skips inert box and child, got %d", got)
	}

	if got := NextFocus(n, 3, 1); got != 3 {
		t.Fatalf("wraps to the only live node, got %d", got)
	}

	if got := NextFocus(n, -1, -1); got != 3 {
		t.Fatalf("backwards, got %d", got)
	}
}

// A pointer timeline on the x axis must stay idle while the pointer is below
// the box, even when its x lines up with the box.
func TestPointerTimelineNeedsPointerOverBox(t *testing.T) {
	nodes := []Node{
		{ID: "ptr", Parent: -1, Rect: Rect{Y: 100, W: 200, H: 50}, Timelines: []Timeline{{Name: "--p", Kind: KindPointer, Axis: AxisX}}},
	}
	h := NewHost(nodes, nil, nil, nil, nil)
	v := Viewport{H: 500, W: 800, ContentH: 2000, ContentW: 800}

	over := h.Compute(v, 100, 120).Timelines[tlKey{0, "--p"}]
	if !over.Inside {
		t.Fatal("pointer over the box must drive the timeline")
	}

	below := h.Compute(v, 100, 300).Timelines[tlKey{0, "--p"}]
	if below.Inside || below.Progress != 0 {
		t.Fatalf("pointer below the box must leave it idle, got inside=%v progress=%v", below.Inside, below.Progress)
	}
}

// A click on the auto panel outlines it. A click on the inert panel reaches
// nothing, so it outlines nothing.
func TestClickOutlinesTheBoxItHit(t *testing.T) {
	nodes := []Node{
		{ID: "inert-panel", Parent: -1, Rect: Rect{Y: 0, W: 300, H: 60}, Inert: true},
		{ID: "auto-panel", Parent: -1, Rect: Rect{Y: 100, W: 300, H: 60}},
		{ID: "plain-button", Parent: -1, Rect: Rect{Y: 200, W: 300, H: 60}},
	}
	h := NewHost(nodes, nil, nil, nil, nil)
	base := map[string]string{
		"inert-panel":  "color:#000;interactivity:inert",
		"auto-panel":   "color:#111;interactivity:auto",
		"plain-button": "color:#222",
	}
	v := Viewport{H: 500, W: 800, ContentH: 500, ContentW: 800}

	if got := h.Styles(h.Compute(v, 10, 130), base); len(got) != 0 {
		t.Fatalf("no click yet, no patches: %+v", got)
	}

	h.SetClicked(HitTest(h.Nodes, 10, 130)) // auto panel
	got := h.Styles(h.Compute(v, 10, 130), base)

	if len(got) != 1 || got[0].ID != "auto-panel" || !strings.Contains(got[0].Style, "outline:3px solid "+colorClick) {
		t.Fatalf("auto panel must be outlined: %+v", got)
	}

	// A box that never declares interactivity is clicked but not outlined.
	h.SetClicked(HitTest(h.Nodes, 10, 230)) // plain-button
	if got = h.Styles(h.Compute(v, 10, 230), base); len(got) != 0 {
		t.Fatalf("a box without interactivity must not be outlined: %+v", got)
	}

	h.SetClicked(HitTest(h.Nodes, 10, 30)) // inert panel: reaches nothing
	if got = h.Styles(h.Compute(v, 10, 30), base); len(got) != 0 {
		t.Fatalf("a click on the inert panel must outline nothing: %+v", got)
	}
}

func TestHostDrivesFrame(t *testing.T) {
	nodes := []Node{
		{ID: "bar", Parent: -1, Rect: Rect{W: 600, H: 12}, Timelines: []Timeline{{Name: "--page", Kind: KindScroll, Axis: AxisY}}},
		{ID: "card", Parent: -1, Rect: Rect{Y: 1000, W: 300, H: 100}, Timelines: []Timeline{{Name: "--c", Kind: KindView, Axis: AxisY}}},
		{ID: "ptr", Parent: -1, Rect: Rect{Y: 100, W: 200, H: 50}, Timelines: []Timeline{{Name: "--p", Kind: KindPointer, Axis: AxisX}}},
	}
	trig := []OwnedTrigger{{Owner: 2, TimelineTrigger: TimelineTrigger{
		Source:          "--page",
		ActivationStart: RangeBound{HasPercent: true, Percent: 50},
		ActivationEnd:   RangeBound{HasPercent: true, Percent: 100},
	}}}
	h := NewHost(nodes, trig, nil, nil, nil)
	base := map[string]string{"bar": "width:600px;height:12px"}
	v := Viewport{ScrollY: 1000, H: 500, W: 800, ContentH: 2000, ContentW: 800}

	f := h.Compute(v, 100, 120)
	near(t, f.Timelines[tlKey{0, "--page"}].Progress, 1000.0/1500)
	near(t, f.Timelines[tlKey{2, "--p"}].Progress, 0.5)

	if f.TriggerOn[0] {
		t.Fatal("--page is declared on a non-ancestor; without timeline-scope the trigger stays off")
	}

	// Hoisting --page to a common ancestor makes the trigger find it.
	h.Nodes[0].Parent = 3
	h.Nodes[2].Parent = 3
	h.Nodes = append(h.Nodes, Node{ID: "root", Parent: -1, Scope: Scope{Kind: ScopeAll}})

	if f = h.Compute(v, 100, 120); !f.TriggerOn[0] {
		t.Fatal("scoped --page is found; 1000 of 1500 is past the 50% activation")
	}

	patches := h.Styles(f, base)
	if len(patches) != 3 || patches[0].ID != "bar" || patches[0].Style != "width:600px;height:12px;width:400.0px" {
		t.Fatalf("patches: %+v", patches)
	}
}

func TestApplierDiffResets(t *testing.T) {
	a := NewApplier(map[string]string{"x": "base"})

	if got := a.Diff([]StylePatch{{"x", "base;w:1"}}); len(got) != 1 {
		t.Fatal("first write")
	}

	if got := a.Diff([]StylePatch{{"x", "base;w:1"}}); len(got) != 0 {
		t.Fatal("unchanged writes nothing")
	}

	if got := a.Diff(nil); len(got) != 1 || got[0].Style != "base" {
		t.Fatalf("driven value gone returns to base: %+v", got)
	}
}

// A progress bar fixed to the window runs to the window's far edge at the end
// of the page, whatever the window width. A page-scroll box that is not fixed
// keeps the width its own style gives it.
func TestFixedBarSpansTheWindow(t *testing.T) {
	nodes := []Node{
		{ID: "bar", Parent: -1, Rect: Rect{W: 600, H: 12}, Timelines: []Timeline{{Name: "--a", Kind: KindScroll, Axis: AxisY}}},
		{ID: "box", Parent: -1, Rect: Rect{Y: 20, W: 300, H: 30}, Timelines: []Timeline{{Name: "--b", Kind: KindScroll, Axis: AxisY}}},
	}
	h := NewHost(nodes, nil, nil, nil, nil)
	base := map[string]string{
		"bar": "position:fixed;top:0;width:600px;height:12px",
		"box": "width:300px;height:30px",
	}

	for _, width := range []float64{800, 1200, 1908} {
		v := Viewport{ScrollY: 1500, H: 500, W: width, ContentH: 2000, ContentW: width} // the very end
		styles := map[string]string{}

		for _, p := range h.Styles(h.Compute(v, 0, 0), base) {
			styles[p.ID] = p.Style
		}

		if want := fmt.Sprintf("width:%.1fpx", width); !strings.HasSuffix(styles["bar"], want) {
			t.Fatalf("window %v: bar style %q, want it to end with %q", width, styles["bar"], want)
		}

		if !strings.HasSuffix(styles["box"], "width:300.0px") {
			t.Fatalf("window %v: a box that is not fixed must keep its width, got %q", width, styles["box"])
		}
	}
}
