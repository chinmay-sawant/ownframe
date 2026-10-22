package page

import (
	"context"
	"testing"
)

// TestPatchElementEditsOneBox checks that an element patch edits the cached
// tree, copies the element's operations into the retained display, and marks
// the element's region dirty.
func TestPatchElementEditsOneBox(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML: `<div id="box" class="a">hello</div><div id="other">world</div>`,
		Theme: `#box{width:80px;height:40px;background:#112233}` +
			`#box.on{background:#445566}`,
		Width:       300,
		Height:      200,
		RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	before := p.Stats().Repaints
	if p.boxByID("box").ID == "" {
		t.Fatal("no box")
	}
	if p.display == nil {
		t.Fatal("the page has no display list")
	}
	if fill := p.boxFill("box"); fill == nil {
		t.Fatal("the box owns no fill")
	} else if fill.R > 0.2 {
		t.Fatalf("the fill starts at %.2f, want the base colour", fill.R)
	}

	mustPatch(t, p.PatchElement(ctx, "box", ElementEdit{Attribute: "class", Value: "a on"}))
	if p.Stats().Repaints != before+1 {
		t.Fatalf("repaints = %d, want one more", p.Stats().Repaints)
	}
	fill := p.boxFill("box")
	if fill == nil {
		t.Fatal("the box owns no fill")
	}
	if fill.R < 0.2 {
		t.Fatalf("the fill is still %.2f, want the patched colour", fill.R)
	}
	if fill.R < 0.2 {
		t.Fatalf("the fill is still %.2f, want the patched colour", fill.R)
	}
}

// boxFill returns the first fill operation the box owns.
func (p *Page) boxFill(id string) *DisplayOp {
	for _, b := range p.display.Boxes {
		if b.ID != id {
			continue
		}
		for i := range p.display.Ops {
			for _, own := range b.OperationIDs {
				if p.display.Ops[i].ID == own && p.display.Ops[i].Kind == DisplayOpFillRect {
					return &p.display.Ops[i]
				}
			}
		}
	}
	return nil
}
