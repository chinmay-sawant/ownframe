package main

import (
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestWindowMatchesFullRows is the 1.0 list-windowing check: a SetScrollWindow
// slice matches full-document row geometry at several offsets.
func TestWindowMatchesFullRows(t *testing.T) {
	full, _, err := fixture(false)
	if err != nil {
		t.Fatal(err)
	}
	if err := redraw(full); err != nil {
		t.Fatal(err)
	}
	window, v, err := fixture(false)
	if err != nil {
		t.Fatal(err)
	}
	installWindow(window, v)
	find := func(p *ownframe.Page, id string) ownframe.Box {
		for _, b := range p.Boxes() {
			if b.ID == id {
				return b
			}
		}
		t.Fatalf("box %s missing", id)
		return ownframe.Box{}
	}
	for i, offset := range []int{0, 20000, 37000} {
		window.SetScrollOffset(0, offset)
		if err := redraw(window); err != nil {
			t.Fatal(err)
		}
		id := fmt.Sprintf("row-%d", []int{0, 280, 479}[i])
		a, b := find(full, id), find(window, id)
		if a.X != b.X || a.Y != b.Y || a.W != b.W || a.H != b.H || a.Text != b.Text {
			t.Fatalf("row %s changed geometry or text", id)
		}
		if find(window, "grid").H != find(full, "grid").H {
			t.Fatal("grid height changed")
		}
		if len(window.Probe().Display().Ops) >= 1500 {
			t.Fatal("window retained too many operations")
		}
	}
}
