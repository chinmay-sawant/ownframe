package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3SelectOpen pins Select closed->open (same value/items)
// as FULL by design: the listbox panel mounts, so the root child count
// changes 1->2 and the ui diff bails out to exactly one reparse.
func TestProbePatchFix3SelectOpen(t *testing.T) {
	t.Parallel()
	items := []chadcn.SelectItem{
		{Value: "a", Label: "Alpha"},
		{Value: "b", Label: "Beta"},
	}
	v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-3", Open: false, Value: "a", Items: items,
		}),
	))
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, ui.Div().ID("probe-root").Children(
		chadcn.Select(chadcn.SelectOptions{
			ID: "sel-3", Open: true, Value: "a", Items: items,
		}),
	)); err != nil {
		t.Fatal(err)
	}
	s1 := v.Page.Probe().Stats()
	if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
		t.Fatalf("generation = %d, want %d", got, want)
	}
	// Pinned FULL: the panel child mounts on open, so the diff cannot patch.
	if got, want := s1.Parses, s0.Parses+1; got != want {
		t.Fatalf("parses = %d, want %d (expected the full-render fallback)", got, want)
	}
}
