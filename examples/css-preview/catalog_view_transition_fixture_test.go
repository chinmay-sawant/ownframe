package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func catalogViewDisplay(t *testing.T, source string) *ownframe.Display {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{HTML: source, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p.Probe().Display()
}

func catalogViewIndex(d *ownframe.Display, id string) int {
	for i := range d.Boxes {
		if d.Boxes[i].ID == id {
			return i
		}
	}
	return -1
}
