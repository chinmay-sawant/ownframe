package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func fixturePage(t *testing.T) (*ownframe.Page, *demo, *view) {
	t.Helper()
	ctx := context.Background()
	v := &view{Started: map[string]bool{}}
	p, err := openCatalog(ctx, catalogOpen{
		Data: v, Width: 960, Height: 1000, RetainState: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Testing().Display() == nil {
		t.Fatal("CSS preview fixture fell back from the display-list path")
	}
	v.Catalog = catalogRows(nil)
	v.Rows = v.Catalog
	if err = redrawCatalog(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	if p.Testing().Display() == nil {
		t.Fatal("CSS preview catalog fell back from the display-list path")
	}
	v.Catalog = catalogRows(p)
	v.Rows = v.Catalog
	v.Readout = describe(p)
	if err = redrawCatalog(ctx, p, v); err != nil {
		t.Fatal(err)
	}
	d := newDemo(p, v)
	p.Handle(ownframe.Handlers{Click: d.click, ClickPaint: d.clickPaint})
	p.Testing().SetAfterRedraw(func() { d.afterRedraw(context.Background()) })
	return p, d, v
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
