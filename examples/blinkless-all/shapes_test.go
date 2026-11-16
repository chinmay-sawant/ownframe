package main

import (
	"context"
	"html"
	"math"
	"strings"
	"testing"
)

// shapesDemoPieces lays out one tab demo's markup and counts border pieces per id.
func shapesPiecesOf(t *testing.T, markup string) map[string][]float64 {
	t.Helper()

	d, err := shapesLayout(context.Background(), markup)
	if err != nil {
		t.Fatal(err)
	}

	out := map[string][]float64{}

	for id, ops := range shapesLines(d) {
		for _, op := range ops {
			out[id] = append(out[id], math.Max(op.W, op.H)/0.75)
		}
	}

	return out
}

// TestShapesClipBoxesPaintCutEdges: every clip box paints one piece more than
// the uncut box on each edge it cuts, and only there.
func TestShapesClipBoxesPaintCutEdges(t *testing.T) {
	cut := shapesBox
	boxes := cut("bc-plain", "", "")
	boxes += cut("bc-all", "border-clip:24px 12px", "")
	boxes += cut("bc-top", "border-top-clip:20px 10px", "") + cut("bc-right", "border-right-clip:20px 10px", "")
	boxes += cut("bc-bottom", "border-bottom-clip:20px 10px", "") + cut("bc-left", "border-left-clip:20px 10px", "")
	boxes += cut("bc-block", "border-block-clip:24px 12px", "")
	boxes += cut("bc-bstart", "border-block-start-clip:20px 10px", "") + cut("bc-bend", "border-block-end-clip:20px 10px", "")
	boxes += cut("bc-inline", "border-inline-clip:24px 12px", "")
	boxes += cut("bc-istart", "border-inline-start-clip:20px 10px", "") + cut("bc-iend", "border-inline-end-clip:20px 10px", "")

	got := shapesPiecesOf(t, boxes)

	for id, want := range map[string]int{
		"bc-plain": 4, "bc-all": 8, "bc-top": 5, "bc-right": 5, "bc-bottom": 5, "bc-left": 5,
		"bc-block": 6, "bc-bstart": 5, "bc-bend": 5, "bc-inline": 6, "bc-istart": 5, "bc-iend": 5,
	} {
		if len(got[id]) != want {
			t.Errorf("%s paints %d line pieces, want %d", id, len(got[id]), want)
		}
	}
}

// TestShapesBorderLimitShortensTop: limit left 60px keeps 60px of the top edge.
func TestShapesBorderLimitShortensTop(t *testing.T) {
	got := shapesPiecesOf(t, shapesBox("bl-left", "border-limit:left 60px", ""))

	found := false

	for _, length := range got["bl-left"] {
		if math.Abs(length-60) < 0.1 {
			found = true
		}

		if length > 60.1 && length > 50 {
			t.Errorf("an edge is %.1fpx, longer than the 60px limit", length)
		}
	}

	if !found {
		t.Errorf("no 60px horizontal piece in %v", got["bl-left"])
	}
}

// TestShapesColumnRuleEdges: around and all paint three vertical rules, the
// gap rule plus both content edges; between and normal paint one.
func TestShapesColumnRuleEdges(t *testing.T) {
	d, err := shapesLayout(context.Background(),
		shapesColumns("between")+shapesColumns("around")+shapesColumns("all")+shapesColumns("normal"))
	if err != nil {
		t.Fatal(err)
	}

	lines := shapesLines(d)

	for id, want := range map[string]int{"cr-between": 1, "cr-around": 3, "cr-all": 3, "cr-normal": 1} {
		got := 0

		for _, op := range lines[id] {
			if op.W == 0 && op.H > 0 {
				got++
			}
		}

		if got != want {
			t.Errorf("%s paints %d vertical rules, want %d", id, got, want)
		}
	}
}

// TestShapesStoredOnlyValuesAreExported: the stored-only demos show the value
// and say it is stored only.
func TestShapesStoredOnlyValuesAreExported(t *testing.T) {
	for name, want := range map[string]string{
		"border-boundary": `"parent"`, "border-shape": `"ellipse(50% 50%)"`,
		"text-fit": `"grow consistent 150%"`, "float-defer": `"1"`, "shape-image-threshold": `"0.5"`,
	} {
		var demo string

		for _, p := range shapesTab().Props {
			if p.Name == name {
				demo = html.UnescapeString(p.Demo())
			}
		}

		if !strings.Contains(demo, name+" stored as "+want+", stored only") {
			t.Errorf("%s demo lacks its stored-only readout", name)
		}
	}
}

// TestShapesEveryClipSiblingHasABox: each clip name the tab claims appears in
// a box caption inside its row's demo.
func TestShapesEveryClipSiblingHasABox(t *testing.T) {
	tab := shapesTab()
	if tab.Count() != 18 {
		t.Fatalf("shapes covers %d, want 18", tab.Count())
	}

	for _, p := range tab.Props {
		demo := p.Demo()

		for _, name := range p.Names() {
			if !strings.Contains(demo, name) {
				t.Errorf("demo of %s never shows %s", p.Name, name)
			}
		}
	}
}
