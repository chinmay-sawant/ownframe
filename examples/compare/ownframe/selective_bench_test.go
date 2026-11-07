package main

import (
	"context"
	"strconv"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/pagetest"
)

// BenchmarkListClickSelective measures the selective click path: the handler
// fingerprints the selection and rebuilds only the header readout, the old
// row, and the new row, reattaching all other rows by pointer so the diff
// fast path skips them. Alternating rows keeps every click a real 3-patch
// change; re-clicking the same row would hit the zero-cost fingerprint path.
func BenchmarkListClickSelective(b *testing.B) {
	ctx := context.Background()
	a, err := newListApp()
	if err != nil {
		b.Fatal(err)
	}
	a.view.Handle(ownframe.Handlers{Click: a.handleClickSelective})
	if err := a.Page().Redraw(ctx); err != nil {
		b.Fatal(err)
	}
	// Prime the row cache before the clock starts so every measured click
	// hits the pointer-identity fast path instead of the one-time full build.
	if err := pagetest.ClickAction(ctx, a.Page(), "row:100"); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	n := 0
	for b.Loop() {
		action := "row:900"
		if n%2 == 1 {
			action = "row:100"
		}
		if err := pagetest.ClickAction(ctx, a.Page(), action); err != nil {
			b.Fatal(err)
		}
		n++
	}
}

// TestSelectiveMatchesNaive drives the selective path and a full-refresh twin
// to the same selections and requires byte-identical page HTML, covering the
// deselect-old/select-new transition and a repeat visit to a cached row.
func TestSelectiveMatchesNaive(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	sel, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	naive, err := newListApp()
	if err != nil {
		t.Fatal(err)
	}
	if err := sel.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := naive.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	for _, want := range []int{100, 900, 100} {
		if err := sel.handleClickSelective(ctx, ownframe.Box{Action: "row:" + strconv.Itoa(want)}); err != nil {
			t.Fatal(err)
		}
		naive.selected = want
		if err := naive.refresh(ctx); err != nil {
			t.Fatal(err)
		}
		if got, wantHTML := sel.Page().HTML(), naive.Page().HTML(); got != wantHTML {
			t.Fatalf("selective HTML differs from naive refresh at selection %d", want)
		}
	}
}
