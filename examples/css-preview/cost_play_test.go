package main

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// costPlayAll clicks the Play all button, the way a person does. The
// click redraws the page with the started class, which is what binds the
// animations, so the harness must not call playAll directly.
func costPlayAll(ctx context.Context, p *ownframe.Page, tb testing.TB) {
	tb.Helper()
	box, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		tb.Fatalf("missing %s", playAllID)
	}
	if err := p.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		tb.Fatal(err)
	}
}
