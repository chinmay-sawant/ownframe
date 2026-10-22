package ownframe

import (
	"bytes"
	"context"
	"image/png"
	"testing"
)

// TestVisualRepaintFullVsPartial compares decoded pixels of an incremental
// repaint with a full Redraw.
func TestVisualRepaintFullVsPartial(t *testing.T) {
	t.Parallel()

	for _, c := range repaintCases() {
		c := c
		t.Run(c.name+"/visual", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			inc := c.newPage(t)
			full := c.newPage(t)

			c.interact(t, inc, ctx)
			c.interact(t, full, ctx)
			if err := full.Redraw(ctx); err != nil {
				t.Fatal(err)
			}

			incPNG := inc.Testing().PNG()
			fullPNG := full.Testing().PNG()
			if bytes.Equal(incPNG, fullPNG) {
				return
			}

			a, err := png.Decode(bytes.NewReader(incPNG))
			if err != nil {
				t.Fatal(err)
			}
			b, err := png.Decode(bytes.NewReader(fullPNG))
			if err != nil {
				t.Fatal(err)
			}
			if a.Bounds() != b.Bounds() {
				t.Fatalf("bounds differ: %v vs %v", a.Bounds(), b.Bounds())
			}

			diff := 0
			for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
				for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
					if a.At(x, y) != b.At(x, y) {
						diff++
					}
				}
			}
			if diff != 0 {
				t.Fatalf("visual diff: %d pixels", diff)
			}
		})
	}
}
