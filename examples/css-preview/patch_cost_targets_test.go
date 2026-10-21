package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/html"
	"github.com/chinmay-sawant/blinkless/layout"
)

// TestPatchTargetsReport prints what the engine does with each element of
// the fixture: the element it splices, and the reason it refuses the rest.
// The report quotes this table.
func TestPatchTargetsReport(t *testing.T) {
	ctx := context.Background()
	body := patchCostSource(t)
	sheet, err := css.Parse(string(patchCostTheme(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range patchCostTargets {
		tree, err := html.Parse(body)
		if err != nil {
			t.Fatal(err)
		}
		styled, err := css.Apply(ctx, tree, css.Options{WidthPx: patchCostWidth,
			HeightPx: patchCostHeight, Extra: []*css.Sheet{sheet}})
		if err != nil {
			t.Fatal(err)
		}
		display, err := layout.DisplayListOptions(ctx, styled,
			layout.Options{RetainState: true})
		if err != nil {
			t.Fatal(err)
		}
		class, ok := patchCostProbeClass[id]
		if !ok {
			class = "zz-probe"
		}
		if !tree.SetAttribute(id, "class", class) {
			t.Logf("%-32s no element", id)
			continue
		}
		styled2, err := css.Relayout(ctx, styled, patchCostWidth, patchCostHeight, "", "", "")
		if err != nil {
			t.Fatal(err)
		}
		patch, err := layout.RelayoutElement(ctx, styled2, display, id, layout.Options{})
		if err != nil {
			t.Logf("%-32s refused: %v", id, err)
			continue
		}
		t.Logf("%-32s spliced: %d operations, old %v new %v",
			id, len(patch.Ops), patch.Old, patch.New)
	}
}
