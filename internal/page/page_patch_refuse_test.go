package page

import (
	"context"
	"errors"
	"testing"
)

// TestPatchElementRefusedLeavesTheTreeAlone checks that a refused patch
// puts the tree back: the page shows the state it had before the edit.
func TestPatchElementRefusedLeavesTheTreeAlone(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#s{font-family:sans-serif}</style>`+
		`<div id="d" style="width:200px"><span id="s">abc</span></div>`)

	text := "a much longer text that resizes the span"
	mustRefuse(t, p.PatchElement(ctx, "s", ElementEdit{Text: &text}))
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for _, op := range p.draw.Display.Ops {
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
	mustRefuse(t, p.PatchElement(context.Background(), "p", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementRefusesAnEmptyID checks that a patch without an element
// id is refused.
func TestPatchElementRefusesAnEmptyID(t *testing.T) {
	p := patchPage(t, `<p id="p">x</p>`)
	mustRefuse(t, p.PatchElement(context.Background(), "", ElementEdit{Attribute: "class", Value: "on"}))
}

func TestPatchElementRefusesANilPage(t *testing.T) {
	var p *Page
	err := p.PatchElement(context.Background(), "p", ElementEdit{Attribute: "class", Value: "on"})
	if !errors.Is(err, ErrNilPage) {
		t.Fatalf("err = %v, want ErrNilPage", err)
	}
}
