package main

import (
	"context"
	"fmt"
	"testing"
)

// TestTabsShowFullNames opens several long-named tabs at a wide and a
// narrow window and requires every tab to keep its full measured width:
// no squeezing, no overlap, the whole filename painted, and the close
// button at the tab's right edge.
func TestTabsShowFullNames(t *testing.T) {
	dir := t.TempDir()
	names := []string{
		"android-build-complete.log",
		"android-build-final.log",
		"dino-android-build.log",
	}
	for _, n := range names {
		writeFile(t, dir, n, "hello\nworld\n")
	}

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	for _, n := range names {
		app.openFile(n)
	}

	widths := make(map[string]float64)

	for _, size := range [][2]int{{1500, 900}, {1100, 700}} {
		app.Page().SetSize(size[0], size[1])

		if err := app.Page().Redraw(ctx); err != nil {
			t.Fatal(err)
		}

		boxes := app.Page().Boxes()
		display := app.Page().Probe().Display()
		if display == nil {
			t.Fatal("page fell back to a bitmap")
		}

		var prevX, prevW float64

		for i, name := range names {
			tabID := fmt.Sprintf("tab-%d", i)
			box, ok := boxByID(boxes, tabID)
			if !ok {
				t.Fatalf("size %dx%d: missing box %s", size[0], size[1], tabID)
			}

			if i == 0 && size[0] == 1500 {
				widths[tabID] = box.W
			} else if w, ok := widths[tabID]; ok && box.W < w-1 {
				t.Errorf("size %dx%d: %s is %.1f px wide, want at least %.1f (full name squeezed)",
					size[0], size[1], tabID, box.W, w)
			}

			if i > 0 && box.X < prevX+prevW-1 {
				t.Errorf("size %dx%d: %s at x=%.1f overlaps previous tab ending at %.1f",
					size[0], size[1], tabID, box.X, prevX+prevW)
			}
			prevX, prevW = box.X, box.W

			found := false
			for _, op := range display.Ops {
				if op.Text == name {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("size %dx%d: full name %q not painted", size[0], size[1], name)
			}

			closeBox, ok := boxByID(boxes, fmt.Sprintf("tab-close-%d", i))
			if !ok {
				t.Fatalf("size %dx%d: missing box tab-close-%d", size[0], size[1], i)
			}

			if gap := box.X + box.W - (closeBox.X + closeBox.W); gap < 0 || gap > 8 {
				t.Errorf("size %dx%d: tab-close-%d ends %.1f px before the tab edge, want 0-8 px",
					size[0], size[1], i, gap)
			}
		}
	}
}
