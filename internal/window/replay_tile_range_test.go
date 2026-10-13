package window

import (
	"image"
	"testing"
)

// TestTileRangeMapsADirtyRect pins the tile mapping: a rect maps to the
// range of tiles it intersects, clamped to the content.
func TestTileRangeMapsADirtyRect(t *testing.T) {
	t.Parallel()

	g := newTileGrid(600, 500) // 3 columns, 2 rows

	cases := []struct {
		rect image.Rectangle
		want [4]int
	}{
		{image.Rect(10, 10, 20, 20), [4]int{0, 0, 1, 1}},
		{image.Rect(250, 250, 300, 300), [4]int{0, 0, 2, 2}},
		{image.Rect(255, 10, 257, 20), [4]int{0, 0, 2, 1}},
		{image.Rect(0, 0, 600, 500), [4]int{0, 0, 3, 2}},
		{image.Rect(-40, -40, 10, 10), [4]int{0, 0, 1, 1}},
		{image.Rect(512, 256, 600, 500), [4]int{2, 1, 3, 2}},
		{image.Rect(600, 500, 700, 700), [4]int{0, 0, 0, 0}},
	}

	for _, c := range cases {
		c0, r0, c1, r1 := g.tileRange(c.rect)
		if got := [4]int{c0, r0, c1, r1}; got != c.want {
			t.Fatalf("tileRange(%v) = %v, want %v", c.rect, got, c.want)
		}
	}
}
