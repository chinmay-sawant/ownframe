package main

import (
	"context"
	"strings"
	"testing"
)

// These tests load the real fixture, so they check the engine's lists and
// scroller ids and the host's reading of them together.

func listSession(t *testing.T) *session {
	t.Helper()

	s, err := newSession(context.Background(), fixture(t), 880, 700, nil)
	if err != nil {
		t.Fatal(err)
	}

	return s
}

func timelinesOf(t *testing.T, s *session, id string) []Timeline {
	t.Helper()

	i, ok := s.host.IndexOf(id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	return s.host.Nodes[i].Timelines
}

func valueOf(t *testing.T, s *session, f Frame, id, name string) Value {
	t.Helper()

	i, ok := s.host.IndexOf(id)
	if !ok {
		t.Fatalf("no box %s", id)
	}

	v, ok := f.Timelines[tlKey{i, name}]
	if !ok {
		t.Fatalf("%s has no timeline %s", id, name)
	}

	return v
}

// TestListFixtureMapping checks what the engine exports for each list case in
// section 9 of the fixture and that the adapter keeps all of it.
func TestListFixtureMapping(t *testing.T) {
	s := listSession(t)

	type tl struct {
		name, kind string
		axis       Axis
		scroller   string
		anonymous  bool
	}

	cases := []struct {
		id   string
		want []tl
	}{
		{"list-bar", []tl{{"--list-a", KindScroll, AxisY, "", false}, {"--list-b", KindScroll, AxisX, "", false}}},
		{"list-view-card", []tl{{"--lv-a", KindView, AxisY, "", false}, {"--lv-b", KindView, AxisX, "", false}}},
		{"list-pointer-box", []tl{{"--lp-a", KindPointer, AxisX, "", false}, {"--lp-b", KindPointer, AxisY, "", false}}},
		// none keeps its slot: --ln-b takes the second axis, y, and nothing is
		// exported for the none item.
		{"list-none-slot", []tl{{"--ln-b", KindScroll, AxisY, "", false}}},
		{"inner-scroller", []tl{{"--inner", KindScroll, AxisY, "inner-scroller", false}}},
		{"inner-card", []tl{{"--ic", KindView, AxisY, "inner-scroller", false}}},
		{"anon-card", []tl{{"--anon", KindView, AxisY, "", true}}},
	}

	for _, test := range cases {
		got := timelinesOf(t, s, test.id)
		if len(got) != len(test.want) {
			t.Fatalf("%s: got %d timelines %+v, want %d", test.id, len(got), got, len(test.want))
		}

		for idx, want := range test.want {
			g := got[idx]
			if g.Name != want.name || g.Kind != want.kind || g.Axis != want.axis ||
				g.ScrollerID != want.scroller || g.Anonymous != want.anonymous {
				t.Fatalf("%s timeline %d = %+v, want %+v", test.id, idx, g, want)
			}
		}
	}

	// The inset list reached the host: --lv-a has 20px and 10px, --lv-b is auto.
	views := timelinesOf(t, s, "list-view-card")
	if views[0].InsetStart < 19.9 || views[0].InsetStart > 20.1 || views[0].InsetEnd < 9.9 || views[0].InsetEnd > 10.1 {
		t.Fatalf("--lv-a insets = %v %v, want 20 10", views[0].InsetStart, views[0].InsetEnd)
	}

	if views[1].InsetStart != 0 || views[1].InsetEnd != 0 {
		t.Fatalf("--lv-b insets = %v %v, want auto", views[1].InsetStart, views[1].InsetEnd)
	}
}

// TestEveryNameInAListIsRead checks both timelines of a list get a value, and
// that a list on an axis the page cannot scroll reads 0.
func TestEveryNameInAListIsRead(t *testing.T) {
	s := listSession(t)
	view := s.viewport(2000)
	f := s.host.Compute(view, 0, 0)

	a := valueOf(t, s, f, "list-bar", "--list-a")
	b := valueOf(t, s, f, "list-bar", "--list-b")

	if a.Progress <= 0 || a.Progress >= 1 {
		t.Fatalf("--list-a is the page y scroll, got %v", a.Progress)
	}

	if b.Progress != 0 {
		t.Fatalf("--list-b is the x axis and the page cannot scroll sideways, got %v", b.Progress)
	}

	la := valueOf(t, s, f, "list-view-card", "--lv-a")
	lb := valueOf(t, s, f, "list-view-card", "--lv-b")

	if la.Inactive || lb.Inactive {
		t.Fatal("page view timelines must be active")
	}

	// Same scroll, different axis: the x axis cannot move, the y axis can.
	if la.Pos == lb.Pos {
		t.Fatalf("--lv-a (y) and --lv-b (x) must read different positions, both %v", la.Pos)
	}
}

// TestDrawnTimelineIsTheFirstOfItsKind checks the style writer draws the first
// scroll timeline of a list, every time, and not a later one.
func TestDrawnTimelineIsTheFirstOfItsKind(t *testing.T) {
	s := listSession(t)
	base := BaseStyles(s.html)

	for run := 0; run < 20; run++ {
		f := s.host.Compute(s.viewport(2000), 0, 0)
		width := ""

		for _, p := range s.host.Styles(f, base) {
			if p.ID == "list-bar" {
				width = p.Style
			}
		}

		// --list-b is 0, so drawing it would give width:0.0px.
		if !strings.Contains(width, "width:") || strings.HasSuffix(width, "width:0.0px") {
			t.Fatalf("run %d: list-bar must follow --list-a, got %q", run, width)
		}
	}
}

// TestNestedScrollerProgress checks the inner box's own scroll timeline and
// the view timeline inside it follow the inner offset, not the page's.
func TestNestedScrollerProgress(t *testing.T) {
	s := listSession(t)

	port, _, ok := s.host.scrollerViewport("inner-scroller")
	if !ok {
		t.Fatal("inner-scroller must be found")
	}

	max := port.ContentH - port.H
	if max <= 0 {
		t.Fatalf("inner-scroller must have content to scroll, content %v window %v", port.ContentH, port.H)
	}

	progress := func(pageY, innerY float64) (own, card float64) {
		s.host.SetScrollerOffset("inner-scroller", 0, innerY)
		f := s.host.Compute(s.viewport(pageY), 0, 0)

		return valueOf(t, s, f, "inner-scroller", "--inner").Progress, valueOf(t, s, f, "inner-card", "--ic").Progress
	}

	own0, card0 := progress(0, 0)
	ownMid, cardMid := progress(0, max/2)
	ownEnd, cardEnd := progress(0, max)

	if own0 != 0 || ownEnd != 1 || ownMid < 0.49 || ownMid > 0.51 {
		t.Fatalf("--inner progress at 0, mid, end = %v %v %v, want 0 0.5 1", own0, ownMid, ownEnd)
	}

	if !(card0 <= cardMid && cardMid <= cardEnd) || card0 == cardEnd {
		t.Fatalf("--ic must advance with the inner offset, got %v %v %v", card0, cardMid, cardEnd)
	}

	// The page scroll must not move a timeline that reads the inner box.
	ownPage, cardPage := progress(3000, max/2)
	if ownPage != ownMid || cardPage != cardMid {
		t.Fatalf("page scroll changed the inner timelines: %v %v vs %v %v", ownPage, cardPage, ownMid, cardMid)
	}
}

// TestNestedScrollerFill checks the scroller's own timeline fills its
// background and leaves its width alone, so the box does not collapse.
func TestNestedScrollerFill(t *testing.T) {
	s := listSession(t)
	s.host.SetScrollerOffset("inner-scroller", 0, 1e6) // clamps to the end
	f := s.host.Compute(s.viewport(0), 0, 0)

	for _, p := range s.host.Styles(f, BaseStyles(s.html)) {
		if p.ID != "inner-scroller" {
			continue
		}

		if !strings.Contains(p.Style, "linear-gradient(90deg,"+colorNested) || strings.Contains(p.Style, ";width:") {
			t.Fatalf("inner-scroller style = %q", p.Style)
		}

		return
	}

	t.Fatal("inner-scroller got no patch")
}

// TestAnonymousScrollerIsInactive checks a view timeline inside a scroll
// container with no id reads 0, is flagged inactive, and draws nothing.
func TestAnonymousScrollerIsInactive(t *testing.T) {
	s := listSession(t)
	f := s.host.Compute(s.viewport(1000), 0, 0)
	v := valueOf(t, s, f, "anon-card", "--anon")

	if !v.Inactive || v.Progress != 0 {
		t.Fatalf("anon-card value = %+v, want inactive and 0", v)
	}

	for _, p := range s.host.Styles(f, BaseStyles(s.html)) {
		if p.ID == "anon-card" {
			t.Fatalf("an inactive timeline must draw nothing, got %q", p.Style)
		}
	}
}

// TestUnknownScrollerIsInactive checks a timeline that names a scroller the
// page does not have stays inactive instead of reading the page.
func TestUnknownScrollerIsInactive(t *testing.T) {
	nodes := []Node{
		{ID: "a", Parent: -1, Rect: Rect{W: 100, H: 100},
			Timelines: []Timeline{{Name: "--x", Kind: KindScroll, Axis: AxisY, ScrollerID: "missing"}}},
	}
	h := NewHost(nodes, nil, nil, nil, nil)
	f := h.Compute(Viewport{H: 100, W: 100, ContentH: 900, ContentW: 100, ScrollY: 400}, 0, 0)

	if v := f.Timelines[tlKey{0, "--x"}]; !v.Inactive || v.Progress != 0 {
		t.Fatalf("value = %+v, want inactive", v)
	}
}

// TestJKScrollTheContainerUnderThePointer checks the keys that scroll an inner
// box in the live window: J and K move the container under the pointer, stop at
// both ends, do nothing when the pointer is elsewhere, and the content shows it.
func TestJKScrollTheContainerUnderThePointer(t *testing.T) {
	s := listSession(t)
	ctx := context.Background()

	i, ok := s.host.IndexOf("inner-scroller")
	if !ok {
		t.Fatal("no inner-scroller")
	}

	r := s.host.Nodes[i].Rect
	scroll := s.viewport(r.Y - 300).ScrollY
	over := FrameInput{ScrollY: scroll, CursorX: r.X + 20, CursorY: r.Y + 20 - scroll}

	press := func(in FrameInput, key string) FrameResult {
		t.Helper()

		in.Keys = []string{key}

		res, err := s.frame(ctx, in)
		if err != nil {
			t.Fatal(err)
		}

		return res
	}
	offset := func() float64 { return s.host.scrollers["inner-scroller"][1] }

	press(over, "J")
	press(over, "J")

	if offset() != 2*innerScrollStep {
		t.Fatalf("two J presses = %v, want %d", offset(), 2*innerScrollStep)
	}

	// The content moves up by the offset, so the scroll can be seen.
	res := press(over, "K")
	style := ""

	for _, p := range s.host.Styles(res.Frame, BaseStyles(s.html)) {
		if p.ID == "inner-scroller-content" {
			style = p.Style
		}
	}

	if offset() != innerScrollStep || !strings.Contains(style, "margin-top:-40px") {
		t.Fatalf("after K: offset %v, inner-scroller-content style %q", offset(), style)
	}

	// The ends: K stops at 0, J stops at the end of the content.
	for k := 0; k < 5; k++ {
		press(over, "K")
	}

	if offset() != 0 {
		t.Fatalf("K must stop at 0, got %v", offset())
	}

	port, _, _ := s.host.scrollerViewport("inner-scroller")

	for k := 0; k < 40; k++ {
		press(over, "J")
	}

	if offset() != port.ContentH-port.H {
		t.Fatalf("J must stop at %v, got %v", port.ContentH-port.H, offset())
	}

	// Away from the box the keys do nothing.
	before := offset()
	away := over
	away.CursorX = 860

	press(away, "K")

	if offset() != before {
		t.Fatalf("a press away from the box moved it: %v to %v", before, offset())
	}

	// The scroll timeline follows the keys.
	f := s.host.Compute(s.viewport(scroll), over.CursorX, over.CursorY)
	if got := valueOf(t, s, f, "inner-scroller", "--inner").Progress; got != 1 {
		t.Fatalf("--inner at the end = %v, want 1", got)
	}
}

// TestWheelScrollsTheContainerUnderThePointer checks the wheel path the window
// uses: a wheel notch over the inner box scrolls it by one step, at either end
// the box gives the event back so the page scrolls, and away from the box it
// never takes the event.
func TestWheelScrollsTheContainerUnderThePointer(t *testing.T) {
	s := listSession(t)

	i, ok := s.host.IndexOf("inner-scroller")
	if !ok {
		t.Fatal("no inner-scroller")
	}

	r := s.host.Nodes[i].Rect
	px, py := r.X+20, r.Y+20 // page coordinates, as the window hands them over
	offset := func() float64 { return s.host.scrollers["inner-scroller"][1] }

	// Rolling the wheel down reports a negative delta and moves the box down.
	if !s.wheel(px, py, -1) || offset() != wheelStep {
		t.Fatalf("wheel down: took=%v offset=%v, want %d", offset() == wheelStep, offset(), wheelStep)
	}

	if !s.wheel(px, py, 1) || offset() != 0 {
		t.Fatalf("wheel up must move it back to 0, got %v", offset())
	}

	// At the top, rolling up has no room: the box hands the event to the page.
	if s.wheel(px, py, 1) {
		t.Fatal("at the top a wheel up must go to the page")
	}

	port, _, _ := s.host.scrollerViewport("inner-scroller")
	for k := 0; k < 40; k++ {
		s.wheel(px, py, -1)
	}

	if offset() != port.ContentH-port.H {
		t.Fatalf("the wheel must stop at %v, got %v", port.ContentH-port.H, offset())
	}

	// At the bottom, rolling down hands the event to the page.
	if s.wheel(px, py, -1) {
		t.Fatal("at the bottom a wheel down must go to the page")
	}

	// Away from the box the wheel is always the page's.
	if s.wheel(860, py, 1) || s.wheel(860, py, -1) {
		t.Fatal("a wheel away from the box must go to the page")
	}
}
