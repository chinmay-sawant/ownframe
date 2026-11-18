package main

import (
	"context"
	"os"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestBuildMapsEveryField(t *testing.T) {
	d := &layout.Display{Boxes: []layout.Box{
		{ID: "root", ParentIndex: -1, W: 800, H: 2000, Interactivity: "auto",
			TimelineScope: layout.TimelineScope{Names: []string{"--s"}}},
		{ID: "a", ParentIndex: 0, X: 1, Y: 2, W: 3, H: 4, Interactivity: "inert",
			Timelines: []layout.TimelineDef{{Name: "--s", Kind: "view", Axis: "x", InsetStart: 5, InsetEnd: 6}}},
		{ID: "a-child", ParentIndex: 1, Interactivity: "inert"},
		{ID: "b", ParentIndex: 0, Interactivity: "auto",
			TimelineScope: layout.TimelineScope{All: true},
			TimelineTriggers: []layout.TimelineTrigger{{
				Name: "--t", Source: "view()", Sources: []string{"view()"},
				ActivationStart: layout.RangeBound{HasPercent: true, Percent: 20},
				ActivationEnd:   layout.RangeBound{Name: "entry", HasLength: true, Length: 7},
				ActiveStart:     layout.RangeBound{Auto: true},
			}},
			EventTriggers: []layout.EventTrigger{{Name: "--e", Start: []string{"click"}, End: []string{"keydown"}}},
			InterestDelay: layout.InterestDelay{StartMs: 200, EndMs: 500}},
		{ID: "b-child", ParentIndex: 3, InterestDelay: layout.InterestDelay{StartMs: 200, EndMs: 500}},
		{ID: "p", ParentIndex: 0, Timelines: []layout.TimelineDef{{Name: "--ptr", Kind: "pointer", Axis: "y", SubjectID: "p"}}},
	}}

	m := Build(d)

	if m.Nodes[0].Scope.Kind != ScopeNames || m.Nodes[0].Scope.Names[0] != "--s" || m.Nodes[3].Scope.Kind != ScopeAll {
		t.Fatalf("scope: %+v %+v", m.Nodes[0].Scope, m.Nodes[3].Scope)
	}

	tl := m.Nodes[1].Timelines[0]
	if tl.Kind != KindView || tl.Axis != AxisX || tl.InsetStart != 5 || tl.InsetEnd != 6 || m.Nodes[1].Rect != (Rect{1, 2, 3, 4}) {
		t.Fatalf("timeline: %+v", tl)
	}

	if m.Nodes[5].Timelines[0].Kind != KindPointer || m.Nodes[5].Timelines[0].Axis != AxisY {
		t.Fatal("pointer timeline")
	}

	if !m.Nodes[1].Inert || !m.Nodes[2].Inert || m.Nodes[3].Inert {
		t.Fatal("inert follows BoxInert")
	}

	if len(m.Triggers) != 1 || m.Triggers[0].Owner != 3 || m.Triggers[0].Sources[0] != "view()" ||
		m.Triggers[0].ActivationEnd.Name != "entry" || !m.Triggers[0].ActivationEnd.HasLength || !m.Triggers[0].ActiveStart.Auto {
		t.Fatalf("trigger: %+v", m.Triggers)
	}

	if len(m.Events) != 1 || m.Events[0].Owner != 3 || m.Events[0].End[0] != "keydown" {
		t.Fatalf("event: %+v", m.Events)
	}

	// b declares the delay; b-child only inherits it, so it gets no tracker.
	if len(m.Delays) != 1 || m.Delays[3].StartMs != 200 || m.Delays[3].EndMs != 500 {
		t.Fatalf("delays: %+v", m.Delays)
	}
}

func TestAnonymousSources(t *testing.T) {
	nodes := []Node{{ID: "box", Parent: -1, Rect: Rect{Y: 1000, W: 100, H: 100}}}
	v := Viewport{ScrollY: 700, H: 400, W: 800, ContentH: 3000, ContentW: 800}
	mk := func(src string) *Host {
		return NewHost(nodes, []OwnedTrigger{{Owner: 0, TimelineTrigger: TimelineTrigger{
			Sources:         []string{src},
			ActivationStart: RangeBound{HasPercent: true, Percent: 0},
			ActivationEnd:   RangeBound{HasPercent: true, Percent: 100},
		}}}, nil, nil, nil)
	}

	// view(): the box is fully in view at scroll 700, so it is inside cover.
	if !mk("view()").Compute(v, 0, 0).TriggerOn[0] {
		t.Fatal("view() is the box's own view timeline")
	}

	v.ScrollY = 0 // the box has not entered yet: T is negative
	if mk("view()").Compute(v, 0, 0).TriggerOn[0] {
		t.Fatal("view() off before entry")
	}

	if !mk("scroll()").Compute(v, 0, 0).TriggerOn[0] {
		t.Fatal("scroll() is the page scroll: 0 of 2600 is inside 0..100%")
	}

	if mk("none").Compute(v, 0, 0).TriggerOn[0] {
		t.Fatal("none never fires")
	}
}

func fixture(t *testing.T) string {
	t.Helper()

	p := scrollPage()
	if _, err := os.Stat(p); err != nil {
		t.Skip("sibling blinkless fixture not found")
	}

	b, _ := os.ReadFile(p) //nolint:gosec // G304: p is a fixed sibling-checkout fixture path chosen by this test, not user input

	return string(b)
}

func TestRealFixtureWiring(t *testing.T) {
	s, err := newSession(context.Background(), fixture(t), 880, 700, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := s.host
	idx := func(id string) int {
		i, ok := h.IndexOf(id)
		if !ok {
			t.Fatalf("no box %s", id)
		}

		return i
	}

	if h.Nodes[idx("scope-root")].Scope.Kind != ScopeNames {
		t.Fatal("scope-root scope")
	}

	if !h.Nodes[idx("inert-panel")].Inert || h.Nodes[idx("auto-panel")].Inert {
		t.Fatal("inert flags")
	}

	if _, _, ok := ResolveTimeline(h.Nodes, idx("scope-reader"), "--scoped"); !ok {
		t.Fatal("scope-reader must find --scoped through scope-root")
	}

	if len(h.Triggers) < 2 || len(h.Events) < 2 || len(h.interest) < 2 {
		t.Fatalf("triggers %d events %d interest %d", len(h.Triggers), len(h.Events), len(h.interest))
	}

	// Inert panel takes no hit: its point falls through to an ancestor.
	r := h.Nodes[idx("inert-panel")].Rect
	if hit := HitTest(h.Nodes, r.X+5, r.Y+5); hit >= 0 && h.Nodes[hit].ID == "inert-panel" {
		t.Fatal("inert panel was hit")
	}

	r = h.Nodes[idx("auto-panel")].Rect
	if hit := HitTest(h.Nodes, r.X+5, r.Y+5); hit < 0 || h.Nodes[hit].ID != "auto-panel" {
		t.Fatal("auto panel should be hit")
	}

	states := func(scroll float64) map[string]bool {
		f := h.Compute(s.viewport(scroll), 0, 0)
		on := map[string]bool{}

		for i, tr := range h.Triggers {
			on[h.Nodes[tr.Owner].ID] = f.TriggerOn[i]
		}

		return on
	}

	// trigger-longhand reads --scoped (scope-card, y 2047) through scope-root.
	if on := states(1739); !on["trigger-longhand"] || on["trigger-box"] {
		t.Fatalf("near scope-card: %v", on)
	}

	// trigger-box watches its own view timeline (y 4053).
	if on := states(4053 - 300); !on["trigger-box"] || on["trigger-longhand"] {
		t.Fatalf("near trigger-box: %v", on)
	}

	// A click on event-box turns its trigger on.
	r = h.Nodes[idx("event-box")].Rect
	if len(h.Dispatch("click", r.X+5, r.Y+5)) == 0 {
		t.Fatal("click on event-box")
	}
}
