package page

import (
	"context"
	"testing"
)

// TestPatchElementDirtiesTheChangedInk checks that the damage after a patch
// covers the operations that changed, not only the element's own box.
func TestPatchElementDirtiesTheChangedInk(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#d.on{font-size:20px}</style>`+
		`<div id="d" style="width:20px;height:20px;white-space:nowrap">hello world</div>`)

	p.TakeDirty()
	prev := p.draw.Display
	mustPatch(t, p.PatchElement(ctx, "d", ElementEdit{Attribute: "class", Value: "on"}))

	d, full, changed := diffDisplay(prev, p.draw.Display)
	if full || d.n == 0 || changed == 0 {
		t.Fatalf("the patch changed nothing: full %v n %d changed %d", full, d.n, changed)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("the patch left no dirty region")
	}
	for i := 0; i < d.n; i++ {
		if rect.Union(d.rects[i]) != rect {
			t.Fatalf("the dirty region %v misses the changed ink %v", rect, d.rects[i])
		}
	}

	box := p.boxByID("d")
	if rect.Dx() <= int(box.W) {
		t.Fatalf("the dirty region %v is the element's box %v", rect, box)
	}
}

// TestPatchElementRebuildsTheHitList checks that a patched element still
// takes the click.
func TestPatchElementRebuildsTheHitList(t *testing.T) {
	ctx := context.Background()
	p := patchPage(t, `<style>#p{width:100px;height:40px}</style><p id="p">text</p>`)

	mustPatch(t, p.PatchElement(ctx, "p", ElementEdit{Attribute: "class", Value: "on"}))

	box := p.boxByID("p")
	box, ok := p.hitBox(box.X+box.W/2, box.Y+box.H/2)
	if !ok || box.ID != "p" {
		t.Fatalf("the patched element does not take the click: %+v %v", box, ok)
	}
}
