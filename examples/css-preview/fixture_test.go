package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func fixturePage(t *testing.T) (*ownframe.Page, *demo, *view) {
	t.Helper()
	ctx := context.Background()
	v := newView()
	p, err := openCatalog(ctx, catalogOpen{
		Data: v, Width: 960, Height: 1000, RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Probe().Display() == nil {
		t.Fatal("CSS preview fixture fell back from the display-list path")
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	v.applyRows()
	if err = redrawCatalog(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	if p.Probe().Display() == nil {
		t.Fatal("CSS preview catalog fell back from the display-list path")
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	v.applyRows()
	if err = redrawCatalog(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click})
	p.Probe().SetAfterRedraw(func() { d.afterRedraw() })
	return p, d, v
}

// clickID sends a click at the center of the element with the given id.
func clickID(t *testing.T, p *ownframe.Page, id string) {
	t.Helper()
	box, ok := findBox(p.Boxes(), id)
	if !ok {
		t.Fatalf("missing element %q", id)
	}
	if err := p.Click(context.Background(), box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
}

// tickPage advances the automatic driver once.
func tickPage(t *testing.T, p *ownframe.Page) {
	t.Helper()
	if err := p.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func samePaint(a, b []paint) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
