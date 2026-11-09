// Choice patch probes: caption fitting-edit, label fitting-edit, and the
// selected flip take the ui.View.Update patch path instead of a full render.
//
// Isolation outcome (measured, box-stable same-length text swaps "aaa"->"aab"):
// a div patches while a plain span falls back to the full render. The engine
// only patches an element that owns a block box: checkRelayoutBox in
// blinkless internal/layout/relayout_subtree.go refuses any other kind with
// "the element is not a block box". A span nested in a block container stays
// an inline box, so no fitting edit can patch it; a div owns a block box, and
// a span that is a flex item is blockified into one (flex.go allocates flex
// items as boxKindBlock), which is why the Choice check mark and the Button
// label (display:block via .of-button-label) already patch. ChoiceWithID
// therefore gives its label span display:block alongside the part id, gated
// on the host id so Choice without an id renders exactly as before.
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

	t.Run("IsolationSpanFullRender", func(t *testing.T) {
		t.Parallel()
		// The same box-stable edit on a plain inline span cannot patch: the
		// engine refuses the non-block box, so Update falls back to a full
		// render (exactly one reparse, generation still advances by one).
		v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(ui.Span().ID("iso").Text("aaa")))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, ui.Div().ID("probe-root").Children(ui.Span().ID("iso").Text("aab"))); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
			t.Fatalf("generation = %d, want %d", got, want)
		}
		if got, want := s1.Parses, s0.Parses+1; got != want {
			t.Fatalf("parses = %d, want %d (expected the full-render fallback)", got, want)
		}
	})
}
