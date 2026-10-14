package window

import (
	"image"
	"testing"
)

// TestTileRangeTouchesExactlyTheTiles checks the mapping both ways: every
// tile in the span meets the rect, and no tile outside it does.
func TestTileRangeTouchesExactlyTheTiles(t *testing.T) {
	t.Parallel()

	g := newTileGrid(500, 300) // 2 columns, 2 rows
	rects := []image.Rectangle{
		image.Rect(10, 10, 20, 20),
		image.Rect(240, 140, 280, 180),
		image.Rect(0, 0, 500, 300),
		image.Rect(-50, -50, 5, 5),
		image.Rect(499, 299, 520, 320),
	}

	for _, rect := range rects {
		c0, r0, c1, r1 := g.tileRange(rect)

		for row := 0; row < g.rows; row++ {
			for col := 0; col < g.cols; col++ {
				in := col >= c0 && col < c1 && row >= r0 && row < r1
				if meets := !rect.Intersect(g.rectAt(col, row)).Empty(); in != meets {
					t.Fatalf("rect %v tile %d,%d: in span = %v, meets = %v", rect, col, row, in, meets)
				}
			}
		}
	}
}

// TestWithinTilesNeedsTheChromeInside pins the chrome-restore guard: chrome
// that reaches past the content forces a full frame.
func TestWithinTilesNeedsTheChromeInside(t *testing.T) {
	t.Parallel()

	g := newTileGrid(100, 100)
	if !withinTiles(g, []image.Rectangle{image.Rect(0, 0, 100, 100)}) {
		t.Fatal("a rect inside the content was rejected")
	}

	if withinTiles(g, []image.Rectangle{image.Rect(0, 0, 101, 10)}) {
		t.Fatal("chrome past the content was accepted")
	}
}
