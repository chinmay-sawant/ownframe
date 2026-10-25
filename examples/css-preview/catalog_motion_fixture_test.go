package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// motionPage builds a page from raw HTML and CSS, draws it, and returns it.
// The CSS controls the animation; no Go binds a player.
func motionPage(t *testing.T, html, css string) *ownframe.Page {
	t.Helper()
	p, err := ownframe.New(ownframe.Config{HTML: html, Theme: css, Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	return p
}

// targetOp returns the first display operation the element with id owns.
func targetOp(t *testing.T, p *ownframe.Page, id string) *ownframe.DisplayOp {
	t.Helper()
	box, ok := findBox(p.Boxes(), id)
	if !ok {
		t.Fatalf("missing target box %q", id)
	}
	indices := cssmotion.OperationIndices(p.Probe().Display(), box)
	if len(indices) == 0 {
		t.Fatalf("target %q owns no display operation", id)
	}
	return &p.Probe().Display().Ops[indices[0]]
}

// targetPaints returns the paints of every operation the element owns.
func targetPaints(t *testing.T, p *ownframe.Page, id string) []paint {
	t.Helper()
	box, ok := findBox(p.Boxes(), id)
	if !ok {
		t.Fatalf("missing target box %q", id)
	}
	return cssmotion.Capture(p.Probe().Display(), box)
}

// xformX is the painted x translation of an operation in CSS pixels. A
// transform-list keyframe writes the matrix and leaves the op's laid-out X
// in place.
func xformX(op *ownframe.DisplayOp, pointsPerPixel float64) float64 {
	if pointsPerPixel <= 0 {
		pointsPerPixel = 1
	}
	return op.Transform().E / pointsPerPixel
}
