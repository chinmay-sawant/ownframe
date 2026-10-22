package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestRowHoverPaintsWithoutLayout(t *testing.T) {
	p, _, err := fixture(true)
	if err != nil {
		t.Fatal(err)
	}
	if err = installHover(p); err != nil {
		t.Fatal(err)
	}
	if err = redraw(p); err != nil {
		t.Fatal(err)
	}
	generation := p.Generation()
	rows := map[string]ownframe.Box{}
	for _, b := range p.Boxes() {
		rows[b.ID] = b
	}
	c := rowPaintCache{rows: map[string]*ownframe.DisplayOp{}}
	c.bind(p)
	for _, id := range []string{"row-1", "row-2"} {
		b := rows[id]
		if err = p.Hover(context.Background(), b.X+20, b.Y+20); err != nil {
			t.Fatal(err)
		}
		op := c.rows[id]
		if op == nil || op.R != 52.0/255 {
			t.Fatalf("row %s not highlighted", id)
		}
		if p.Generation() != generation {
			t.Fatal("hover performed a layout")
		}
		if _, ok := p.TakeDirty(); !ok {
			t.Fatal("hover did not invalidate its pixels")
		}
	}
	if c.rows["row-1"].R != 20.0/255 {
		t.Fatal("previous row not restored")
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err = p.Hover(context.Background(), rows["row-3"].X+20, rows["row-3"].Y+20); err != nil {
		t.Fatal(err)
	}
	c.bind(p)
	if c.rows["row-2"].R != 29.0/255 {
		t.Fatal("hover cache did not rebind after layout")
	}
}
