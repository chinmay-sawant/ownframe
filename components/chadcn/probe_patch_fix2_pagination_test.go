package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// TestProbePatchFix2Pagination covers the Pagination stable-ID fix:
// PaginationOptions.ID gives the nav and every control a stable id
// (ID+"-prev", ID+"-next", ID+"-page-N", ID+"-ellipsis-K") so a changed
// control can be addressed by ui.View.Update. A page move inside a stable
// slot window still takes the full render, because the old current page loses
// its aria-current attribute and the diff cannot patch a removed attribute;
// a slot-window shift mounts different page numbers, so it stays FULL by
// design. The identical rebuild is a NOOP.
func TestProbePatchFix2Pagination(t *testing.T) {
	t.Parallel()

	build := func(page int) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.Pagination(chadcn.PaginationOptions{
			ID: "pp-pag", Page: page, TotalPages: 5,
		}))
	}

	t.Run("IdenticalNoop", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(2))
		expectProbeNoop(t, v, ctx, build(2))
	})

	// SameWindowPageMoveFull stays FULL: pages 2 and 3 mount the same five
	// numbered buttons in the same order with stable ids, and the previous
	// and next buttons keep their handlers and enabled state, but the
	// pp-pag-page-2 button loses its aria-current attribute to
	// pp-pag-page-3. changedAttributes refuses the removal, so the diff
	// reports ok=false and Update reparses.
	t.Run("SameWindowPageMoveFull", func(t *testing.T) {
		t.Parallel()
		v, ctx := mountProbeView(t, build(2))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, build(3)); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		if got := v.Page.Probe().Generation(); got != gen0+1 {
			t.Fatalf("generation = %d, want %d", got, gen0+1)
		}
		if s1.Parses <= s0.Parses {
			t.Fatalf("expected full render: parses %d -> %d", s0.Parses, s1.Parses)
		}
	})

	// SlotWindowShiftFull stays FULL by design: moving from page 5 to page 8
	// of 10 mounts different window numbers (4,5,6 become 7,8,9), so the
	// buttons at the same positions carry different ids and the diff refuses
	// the rename.
	t.Run("SlotWindowShiftFull", func(t *testing.T) {
		t.Parallel()
		buildWide := func(page int) *ui.Element {
			return ui.Div().ID("probe-root").Children(chadcn.Pagination(chadcn.PaginationOptions{
				ID: "pp-pag-wide", Page: page, TotalPages: 10,
			}))
		}
		v, ctx := mountProbeView(t, buildWide(5))
		gen0 := v.Page.Probe().Generation()
		s0 := v.Page.Probe().Stats()
		if err := v.Update(ctx, buildWide(8)); err != nil {
			t.Fatal(err)
		}
		s1 := v.Page.Probe().Stats()
		if got := v.Page.Probe().Generation(); got != gen0+1 {
			t.Fatalf("generation = %d, want %d", got, gen0+1)
		}
		if s1.Parses <= s0.Parses {
			t.Fatalf("expected full render: parses %d -> %d", s0.Parses, s1.Parses)
		}
	})
}
