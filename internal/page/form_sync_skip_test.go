package page

import (
	"context"
	"strings"
	"testing"
)

// TestSyncFormSkipsWithoutControls redraws an input-free page twice: the
// second sync must take the no-controls fast path, which keeps the same
// bookkeeping as the full merge (body recorded, flag cleared, no live
// controls) and still runs the emoji pass.
func TestSyncFormSkipsWithoutControls(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="m">Good 😂 and plain</p>`, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	fs := p.formState()
	if fs == nil {
		t.Fatal("no form state after redraw")
	}

	if len(fs.ByID) != 0 || len(fs.Order) != 0 {
		t.Fatalf("controls = %d/%d, want 0/0", len(fs.ByID), len(fs.Order))
	}

	if p.formDirty {
		t.Fatal("formDirty still set after sync")
	}

	if p.lastSyncBody == "" {
		t.Fatal("lastSyncBody not recorded")
	}

	if strings.Contains(p.source, "data-ownframe-field") {
		t.Fatal("input-free source carries field markup")
	}

	// The emoji pass still runs on the skip path.
	if !strings.Contains(p.source, `src="emoji/1f602"`) {
		t.Fatalf("emoji lost on input-free page: %s", p.source)
	}

	// A focus request for an absent control is consumed, exactly as the
	// full merge does, so it never sticks around for a later template.
	RequestFocus(p, "ghost")

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.ff.hasFocusRequest {
		t.Fatal("focus request left pending on input-free page")
	}

	if got := p.formState(); len(got.ByID) != 0 || len(got.Order) != 0 {
		t.Fatalf("controls = %d/%d after second sync, want 0/0", len(got.ByID), len(got.Order))
	}
}

// TestSyncFormKeepsControlsWithInput proves pages with inputs still run
// the full merge: the live set is populated and a user value survives a
// redraw that re-executes the template.
func TestSyncFormKeepsControlsWithInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p, err := New(Config{HTML: `<input id="t" type="text" value="hi">`, Width: 300, Height: 100})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	fs := p.formState()
	if fs == nil || len(fs.ByID) != 1 || len(fs.Order) != 1 {
		t.Fatalf("live controls not merged: %+v", fs)
	}

	p.SetFormValue("t", "typed")

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if got := p.FormValue("t"); got != "typed" {
		t.Fatalf("value = %q, want %q", got, "typed")
	}

	if !strings.Contains(p.source, "data-ownframe-field") {
		t.Fatal("input page lost its field markup")
	}
}
