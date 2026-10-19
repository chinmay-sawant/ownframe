package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

func motionFixture(t *testing.T, html, css string) (*ownframe.Display, ownframe.Box, []int, []cssmotion.Paint) {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{HTML: html, Theme: css, Width: 300, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	d := p.Display()
	b, ok := findBox(p.Boxes(), "target")
	if d == nil || !ok {
		t.Fatal("fixture did not produce target box and display")
	}
	indices := cssmotion.OperationIndices(d, b)
	initial := cssmotion.Capture(d, b)
	if len(indices) == 0 || len(initial) != len(indices) {
		t.Fatal("fixture did not produce target paint operations")
	}
	return d, b, indices, initial
}
