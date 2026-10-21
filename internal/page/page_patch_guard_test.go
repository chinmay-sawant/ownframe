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
	if p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}) {
		t.Fatal("the patch was not refused")
	}
}

// TestPatchElementRefusesAnEditThatChangesNothing checks that a patch with
// no edit, one without a name, and one on an element the tree does not hold
// are refused.
func TestPatchElementRefusesAnEditThatChangesNothing(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<p id="p">x</p>`)

	if p.PatchElement(ctx, "p") {
		t.Fatal("the empty patch was not refused")
	}
	if p.PatchElement(ctx, "p", ElementEdit{Attribute: "", Value: "on"}) {
		t.Fatal("the patch without a name was not refused")
	}
	if p.PatchElement(ctx, "missing", ElementEdit{Attribute: "class", Value: "on"}) {
		t.Fatal("the patch on a missing element was not refused")
	}
}

// TestPatchElementFollowsTheCacheStateRules checks that the page reads the
// state rules of the cached sheets.
func TestPatchElementFollowsTheCacheStateRules(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:10px;height:10px}</style><p id="p">x</p>`)

	if !p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}) {
		t.Fatal("the patch was refused")
	}
	if p.statePseudos != p.cache.StatePseudos() {
		t.Fatalf("the page reports %+v and the cache %+v", p.statePseudos, p.cache.StatePseudos())
	}
}
