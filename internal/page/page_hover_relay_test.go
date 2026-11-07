package page

import (
	"context"
	"testing"
)

// TestHoverRedrawReusesPatchedCascade pins the single-cascade hover path: a
// hover handler patch relays the cache for the new hover state, so the
// Redraw's styleStaged must reuse that document instead of cascading again.
// Pointer equality proves no second css.Relayout ran.
func TestHoverRedrawReusesPatchedCascade(t *testing.T) {
	source := `<style>.cell{width:40px;height:20px}.cell:hover{outline:2px solid red}#detail{width:200px;height:20px}</style><div id="a" class="cell">a</div><div id="detail">sum</div>`
	p, err := New(Config{HTML: source, Width: 320, Height: 200, RetainState: true})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	// Mirror the reordered setHover: the id moves first, then the handler
	// patch cascades with the new hover state.
	p.setHoverID("a")
	text := "Aisle A"
	if err := p.PatchElement(ctx, "detail", ElementEdit{Text: &text}); err != nil {
		t.Fatal(err)
	}
	relayed := p.cache.Styled()
	doc, err := p.styleStaged(ctx, false, p.source, p.renderState())
	if err != nil {
		t.Fatal(err)
	}
	if doc != relayed {
		t.Fatal("Redraw re-cascaded after a handler patch for the same state; want the relayed document")
	}
}
