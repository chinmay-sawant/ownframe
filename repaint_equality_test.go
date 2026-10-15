package ownframe

import (
	"bytes"
	"context"
	"image/png"
	"testing"
)

// repaintCase is one page whose incremental path must match a full Redraw.
type repaintCase struct {
	name     string
	newPage  func(t *testing.T) *Page
	interact func(t *testing.T, p *Page, ctx context.Context)
}

// TestRepaintMatchesFullRedraw drives one interaction on two pages and
// compares the incremental PNG with the PNG after a full Redraw. This is the
// guard that keeps a partial repaint from being a subtly wrong picture.
func TestRepaintMatchesFullRedraw(t *testing.T) {
	t.Parallel()

	for _, c := range repaintCases() {
		c := c
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			inc := c.newPage(t)
			full := c.newPage(t)

			before := inc.PNG()
			c.interact(t, inc, ctx)
			c.interact(t, full, ctx)
			if err := full.Redraw(ctx); err != nil {
				t.Fatal(err)
			}

			after := inc.PNG()
			if before == nil || after == nil {
				t.Fatal("no PNG to compare")
			}

			beforeImg, err := png.Decode(bytes.NewReader(before))
			if err != nil {
				t.Fatal(err)
			}

			afterImg, err := png.Decode(bytes.NewReader(after))
			if err != nil {
				t.Fatal(err)
			}

			if samePicture(beforeImg, afterImg) {
				t.Fatal("the interaction did not change the picture")
			}

			fullImg, err := png.Decode(bytes.NewReader(full.PNG()))
			if err != nil {
				t.Fatal(err)
			}

			if !samePicture(afterImg, fullImg) {
				t.Fatal("incremental picture differs from a full Redraw")
			}
		})
	}
}

// Helpers newPage, clickID, and boxCenter live in
// repaint_equality_helpers_test.go.
