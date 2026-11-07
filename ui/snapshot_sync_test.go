package ui

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestPatchSnapshotMatchesFullRender pins the View.patch fast-path invariant:
// after a patched update, Page.HTML() must be byte-identical to a full
// renderDocument of the new tree. Covers style edits plus a text edit on a
// box-owning flex-item span through Update (not just the string helpers).
// (A text edit on a box-less plain inline span still refuses at the engine
// gate and takes the full render; that engine text-retarget is a separate
// follow-up, so the fixture keeps its label span a flex item.)
func TestPatchSnapshotMatchesFullRender(t *testing.T) {
	ctx := context.Background()
	mk := func(sel int, label string) *Element {
		rows := make([]*Element, 0, 4)
		for i := range 4 {
			bg := "#ffffff"
			if i == sel {
				bg = "#dbeafe"
			}
			rows = append(rows, Div().ID("snap-row-"+string(rune('a'+i))).
				Attr("data-action", "row").
				BackgroundColor(Hex(bg)).
				Children(Text("r"+string(rune('0'+i)))))
		}
		return Div().ID("snap-root").Children(
			append([]*Element{
				Div().Display(DisplayFlex).ID("snap-label-wrap").Children(
					Text(label).ID("snap-label"),
				),
			}, rows...)...,
		)
	}
	v, err := Mount(ownframe.Config{Title: "snap", Width: 320, Height: 240}, mk(-1, "pick one"))
	if err != nil {
		t.Fatal(err)
	}
	if err := v.Page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	next := mk(2, "picked C")
	before := v.Page.Probe().Stats().Parses
	if err := v.Update(ctx, next); err != nil {
		t.Fatal(err)
	}
	if got := v.Page.Probe().Stats().Parses; got != before {
		t.Fatalf("patched update parsed %d times, want 0 (must stay on the patch path)", got-before)
	}
	r2, html, err := renderDocument("snap", next)
	if err != nil {
		t.Fatal(err)
	}
	putRenderer(r2)
	if v.Page.HTML() != html {
		t.Fatalf("patched HTML differs from full render:\npage: %.200s\nfull: %.200s", v.Page.HTML(), html)
	}
	if html == "" {
		t.Fatal("empty render")
	}
}
