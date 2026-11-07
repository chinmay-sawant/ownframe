package page

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

// Baseline (Nov 2026, 1000-row page): a targeted 3-patch batch costs ~5.1MB
// because the engine escalates to a whole-table relayout. Goal: <100KB per
// click. This fixture is 20 rows, not 1000, so it runs in milliseconds and
// pins the shape of the cost, not the byte count: three row class edits must
// commit through the batch path with one shared style re-resolve and match a
// full-Redraw twin op for op. All three edits target box-owning row divs: a
// text edit on the box-less inner span would refuse at the engine gate
// ("element has no box"), which is correct engine behavior for paint owned
// by the parent, not a batch-path defect.
// Expected direction as fixes land:
//   - row splice keeps the relayout to the touched rows, so B/op falls with
//     row count instead of scaling with the table;
//   - tag interning lowers the edit and string churn inside the batch;
//   - source sync keeps the patched source bookkeeping off the parse path.

const (
	patchBudgetRows   = 20
	patchBudgetWidth  = 320
	patchBudgetHeight = 480
)

// patchBudgetSource renders the small-table fixture: patchBudgetRows
// stable-id rows, each a leaf span inside a classed row. on marks rows that
// already carry the "on" class; cellText overrides a cell's text (used for
// the twin).
func patchBudgetSource(on map[int]bool, cellText map[int]string) string {
	var sb strings.Builder
	sb.WriteString(`<style>.row{width:200px;height:20px}.on{background:#445566}</style><div id="list">`)
	for i := range patchBudgetRows {
		class := "row"
		if on[i] {
			class = "row on"
		}
		text := "item " + strconv.Itoa(i)
		if alt, ok := cellText[i]; ok {
			text = alt
		}
		sb.WriteString(`<div class="` + class + `" id="row-` + strconv.Itoa(i) +
			`"><span id="cell-` + strconv.Itoa(i) + `">` + text + `</span></div>`)
	}
	sb.WriteString(`</div>`)
	return sb.String()
}

// patchBudgetBatch is the pinned 3-edit batch: class toggles on three
// box-owning rows (background-only, geometry-neutral).
func patchBudgetBatch() []ElementPatch {
	return []ElementPatch{
		{ID: "row-3", Edits: []ElementEdit{{Attribute: "class", Value: "row on"}}},
		{ID: "row-7", Edits: []ElementEdit{{Attribute: "class", Value: "row on"}}},
		{ID: "row-11", Edits: []ElementEdit{{Attribute: "class", Value: "row on"}}},
	}
}

// patchBudgetRevert undoes patchBudgetBatch for the alternating benchmark:
// each op must change something, or the batch refuses as a no-op.
func patchBudgetRevert() []ElementPatch {
	return []ElementPatch{
		{ID: "row-3", Edits: []ElementEdit{{Attribute: "class", Value: "row"}}},
		{ID: "row-7", Edits: []ElementEdit{{Attribute: "class", Value: "row"}}},
		{ID: "row-11", Edits: []ElementEdit{{Attribute: "class", Value: "row"}}},
	}
}

func newPatchBudgetPage(t *testing.T, source string) *Page {
	t.Helper()
	p, err := New(Config{HTML: source, Width: patchBudgetWidth, Height: patchBudgetHeight, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// TestPatchBatchSmallTableMatchesFull pins the targeted-patch path: the
// 3-edit batch succeeds and the spliced display matches a full-Redraw twin
// box for box and op for op via assertSameDisplay.
func TestPatchBatchSmallTableMatchesFull(t *testing.T) {
	ctx := context.Background()
	patched := newPatchBudgetPage(t, patchBudgetSource(nil, nil))
	if err := patched.PatchElements(ctx, patchBudgetBatch()); err != nil {
		t.Fatalf("PatchElements = %v, want nil", err)
	}
	full := newPatchBudgetPage(t, patchBudgetSource(map[int]bool{3: true, 7: true, 11: true}, nil))
	assertSameDisplay(t, patched, full)
}

// BenchmarkPatchElementsSmallBatch pins the steady-state cost of the
// targeted 3-edit batch on the small table. Ops alternate between applying
// and reverting so every op changes state (an unchanged batch refuses).
// ReportAllocs surfaces allocs/op and B/op with -benchmem; row splice is
// the fix expected to move B/op, since the relayout scope is the current
// whole-table escalation.
func BenchmarkPatchElementsSmallBatch(b *testing.B) {
	ctx := context.Background()
	p, err := New(Config{HTML: patchBudgetSource(nil, nil), Width: patchBudgetWidth, Height: patchBudgetHeight, RetainState: true})
	if err != nil {
		b.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	on, off := patchBudgetBatch(), patchBudgetRevert()
	b.ReportAllocs()
	for i := 0; b.Loop(); i++ {
		batch := on
		if i%2 == 1 {
			batch = off
		}
		if err := p.PatchElements(ctx, batch); err != nil {
			b.Fatal(err)
		}
	}
}
