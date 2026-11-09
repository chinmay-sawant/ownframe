// Accordion toggle probe: opening a second panel (a1 -> a1+a2) stays FULL.
// The closed panel is lazily unmounted by design (pinned by
// TestAccordionRendersOnlyOpenPanels), so the toggle mounts a new panel node
// (child-count change), which the diff cannot patch. An always-mounted-hidden
// panel would patch but would break that documented mount/unmount semantic.
package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func TestProbePatchFix3Accordion(t *testing.T) {
	t.Parallel()
	build := func(open ...string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Accordion(chadcn.AccordionOptions{
			Open: open,
			Items: []chadcn.AccordionItem{
				{ID: "a1", Title: "One", Content: ui.P().ID("pp-fix3-acc-a1").Text("c1")},
				{ID: "a2", Title: "Two", Content: ui.P().ID("pp-fix3-acc-a2").Text("c2")},
			},
		}))
	}
	v, ctx := mountProbeView(t, build("a1"))
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, build("a1", "a2")); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
		t.Fatalf("generation = %d, want %d", got, want)
	}
	// Pinned FULL: the second panel mounts on open (child-count change), so
	// Update falls back to exactly one full reparse.
	if got, want := s1.Parses, s0.Parses+1; got != want {
		t.Fatalf("parses = %d, want %d (expected the full-render fallback)", got, want)
	}
}
