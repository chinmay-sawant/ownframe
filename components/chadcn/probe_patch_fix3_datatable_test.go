package chadcn_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Pinned FULL: datatable-addrow (2 rows -> 3 rows) cannot take the
// ui.View.Update patch path from data_table.go alone.
func TestProbePatchFix3Datatable(t *testing.T) {
	t.Parallel()
	columns := []chadcn.DataColumn{{Label: "ID"}, {Label: "Total"}}
	mk := func(rows [][]string) *ui.Element {
		return ui.Div().ID("probe-root").Children(chadcn.DataTable(chadcn.DataTableOptions{
			ID: "dt-fix3", Caption: "Orders", Columns: columns, Rows: rows,
		}))
	}
	v, ctx := mountProbeView(t, mk([][]string{{"1", "$9"}, {"2", "$14"}}))
	gen0 := v.Page.Probe().Generation()
	s0 := v.Page.Probe().Stats()
	if err := v.Update(ctx, mk([][]string{{"1", "$9"}, {"2", "$14"}, {"3", "$21"}})); err != nil {
		t.Fatal(err)
	}
	if got, want := v.Page.Probe().Generation(), gen0+1; got != want {
		t.Fatalf("generation = %d, want %d", got, want)
	}
	s1 := v.Page.Probe().Stats()
	// Pinned FULL: the appended row mounts a new tr so the tbody child count
	// grows 2->3 and later box indices shift; ui diff rejects child-count
	// changes before stable ids are compared, so no id scheme in data_table.go
	// can patch this without index-shift/insert support in the ui diff layer
	// (plus engine relayout for the mount).
	if s1.Parses != s0.Parses+1 {
		t.Fatalf("add-row parses = %d, want %d (pinned FULL)", s1.Parses, s0.Parses+1)
	}
}
