package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix3Command pins the command-query narrowed-list case as FULL
// by design: the host owns filtering (Command draws Groups as given), so Query
// ""->"save" with host-narrowed groups unmounts a row (2 rows -> 1), changing
// the section child count, which ui diff refuses. Keeping the filtered rows
// mounted-but-hidden would change semantics pinned by
// TestCommandKeepsHostFilteredGroups (role=option count), so it stays FULL.
func TestProbePatchFix3Command(t *testing.T) {
	t.Parallel()

	full := []chadcn.CommandGroup{{Label: "File", Items: []chadcn.CommandItem{
		{Value: "save", Label: "Save"},
		{Value: "open", Label: "Open"},
	}}}
	narrowed := []chadcn.CommandGroup{{Label: "File", Items: []chadcn.CommandItem{
		{Value: "save", Label: "Save"},
	}}}

	t.Run("SameListKeystrokePatches", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
			chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "", Groups: full, Active: "save"})))
		expectProbePatch(t, v, ctx, ui.Div().ID("probe-root").Children(
			chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "s", Groups: full, Active: "save"})))
	})

	t.Run("HostNarrowedListFull", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, ui.Div().ID("probe-root").Children(
			chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "", Groups: full, Active: "save"})))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		// Pinned FULL: host-narrowed row unmount (2 rows -> 1) changes the
		// group section child count, so ui.View.Update falls back to a full
		// render (exactly one reparse, generation still advances by one).
		if err := v.Update(ctx, ui.Div().ID("probe-root").Children(
			chadcn.Command(chadcn.CommandOptions{ID: "cmd-1", Query: "save", Groups: narrowed, Active: "save"}))); err != nil {
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
