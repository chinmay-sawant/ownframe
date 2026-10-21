package page

import (
	"context"
	"testing"
)

// patchPage builds a page that keeps the layout state an element patch needs.
func patchPage(t *testing.T, source string) *Page {
	t.Helper()

	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return p
}

// TestPatchElementRefusedLeavesTheTreeAlone checks that a refused patch
// puts the tree back: the page shows the state it had before the edit.
func TestPatchElementRefusedLeavesTheTreeAlone(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#s{font-family:sans-serif}</style>`+
		`<div id="d" style="width:200px"><span id="s">abc</span></div>`)

	text := "a much longer text that resizes the span"
	if p.PatchElement(ctx, "s", ElementEdit{Text: &text}) {
		t.Fatal("the patch was not refused")
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for _, op := range p.display.Ops {
		if op.Kind == DisplayOpText && op.W > 60 {
			t.Fatalf("the refused patch changed the text: %+v", op)
		}
	}
}

// TestPatchElementRefusesWithoutACache checks that a page that never drew
// refuses a patch.
func TestPatchElementRefusesWithoutACache(t *testing.T) {
	p, err := New(Config{HTML: `<p id="p">x</p>`, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	if p.PatchElement(context.Background(), "p", ElementEdit{Attribute: "class", Value: "on"}) {
		t.Fatal("the patch was not refused")
	}
}

// TestPatchElementRefusesAnEmptyID checks that a patch without an element
// id is refused.
func TestPatchElementRefusesAnEmptyID(t *testing.T) {
	p := patchPage(t, `<p id="p">x</p>`)
	if p.PatchElement(context.Background(), "", ElementEdit{Attribute: "class", Value: "on"}) {
		t.Fatal("the patch was not refused")
	}
}
