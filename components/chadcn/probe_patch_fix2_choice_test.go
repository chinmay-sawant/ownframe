// Choice patch probes: caption fitting-edit, label fitting-edit, and the
// selected flip take the ui.View.Update patch path instead of a full render.
//
// Isolation outcome (measured, box-stable same-length text swaps "aaa"->"aab"):
// both a div and a plain inline span patch without reparsing. Earlier engine
// builds refused non-block boxes in checkRelayoutBox, so ChoiceWithID gave its
// label span display:block alongside the part id (gated on the host id so
// Choice without an id renders as before). Blinkless 50ad332 now patches
// inline spans too, and ChoiceWithID keeps display:block so its layout stays
// unchanged.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func probeFix2ChoiceRoot(el *ui.Element) *ui.Element {
	return ui.Div().ID("probe-root").Children(el)
}

func TestProbePatchFix2Choice(t *testing.T) {
	t.Parallel()

	t.Run("CaptionFitting", func(t *testing.T) {
		t.Parallel()
		// Same-length digit swap: the caption paragraph keeps its box, and a
		// p owns a block box, so the edit patches.
		v, ctx := mountProbeView(t, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", false)))
		expectProbePatch(t, v, ctx, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 15 days.", false)))
	})

	t.Run("LabelFitting", func(t *testing.T) {
		t.Parallel()
		// Shorter text fits the kept box (the ButtonLabel probe proves a
		// shrinking label patches); the display:block part span owns a block
		// box, so this patches instead of falling back.
		v, ctx := mountProbeView(t, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", false)))
		expectProbePatch(t, v, ctx, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backup", "Keeps 14 days.", false)))
	})

	t.Run("SelectedFlip", func(t *testing.T) {
		t.Parallel()
		// The flip retouches the root button (paint-only style and
		// aria-pressed edits) and fills the check mark, a blockified flex
		// item; the containing flex row keeps its size, so the batch patches.
		v, ctx := mountProbeView(t, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", false)))
		expectProbePatch(t, v, ctx, probeFix2ChoiceRoot(chadcn.ChoiceWithID("choice-1", "Nightly backups", "Keeps 14 days.", true)))
	})

	t.Run("IsolationDivPatches", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(ui.Div().ID("iso").Text("aaa")))
		expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(ui.Div().ID("iso").Text("aab")))
	})

	t.Run("IsolationSpanPatches", func(t *testing.T) {
		t.Parallel()
		// A box-stable text edit on a plain inline span patches without
		// falling back to a full render.
		v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(ui.Span().ID("iso").Text("aaa")))
		expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(ui.Span().ID("iso").Text("aab")))
	})
}
