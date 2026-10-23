package page

import (
	"context"
	"testing"
)

// TestPatchElementRefusesWithoutRetainedState checks that a page without
// the retained layout state refuses a patch.
func TestPatchElementRefusesWithoutRetainedState(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{HTML: `<p id="p">x</p>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	mustRefuse(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementRefusesAnEditThatChangesNothing checks that a patch with
// no edit, one without a name, and one on an element the tree does not hold
// are refused.
func TestPatchElementRefusesAnEditThatChangesNothing(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<p id="p">x</p>`)

	mustRefuse(t, p.PatchElement(ctx, "p"))
	mustRefuse(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "", Value: "on"}))
	mustRefuse(t, p.PatchElement(ctx, "missing", ElementEdit{Attribute: "class", Value: "on"}))
}

// TestPatchElementFollowsTheCacheStateRules checks that the page reads the
// state rules of the cached sheets.
func TestPatchElementFollowsTheCacheStateRules(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:10px;height:10px}</style><p id="p">x</p>`)

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))
	if p.statePseudos != p.cache.StatePseudos() {
		t.Fatalf("the page reports %+v and the cache %+v", p.statePseudos, p.cache.StatePseudos())
	}
}
