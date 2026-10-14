package window

import (
	"image"
	"testing"
)

// TestTileGridKeepsTheContentCost checks the grid partitions the content:
// every tile is clipped at the edge and the tiles add up to the content area,
// the single buffer's cost.
func TestTileGridKeepsTheContentCost(t *testing.T) {
	t.Parallel()

	sizes := []image.Point{{100, 100}, {256, 256}, {600, 500}, {1908, 999}}

	for _, size := range sizes {
		g := newTileGrid(size.X, size.Y)
		if g == nil {
			t.Fatalf("%v: no grid", size)
		}

		want := gridTiles(size.X) * gridTiles(size.Y)
		if len(g.images) != want {
			t.Fatalf("%v: %d tiles, want %d", size, len(g.images), want)
		}

		area := 0

		for row := 0; row < g.rows; row++ {
			for col := 0; col < g.cols; col++ {
				r := g.rectAt(col, row)
				if r.Dx() > tileSize || r.Dy() > tileSize || !r.In(g.content()) {
					t.Fatalf("%v: tile %d,%d = %v", size, col, row, r)
				}

				if b := g.at(col, row).Bounds(); b != image.Rect(0, 0, r.Dx(), r.Dy()) {
					t.Fatalf("%v: tile %d,%d image = %v, want %v", size, col, row, b, r)
				}

				area += r.Dx() * r.Dy()
			}
		}

		if area != size.X*size.Y {
			t.Fatalf("%v: tiles cover %d pixels, want %d", size, area, size.X*size.Y)
		}
	}
}

// TestEdgeTileIsClipped pins the corner tile of a size that is not a
// multiple of the tile size: it is the remainder, not a full tile.
func TestEdgeTileIsClipped(t *testing.T) {
	t.Parallel()

	g := newTileGrid(300, 300) // 2 columns, 2 rows
	if got := g.rectAt(1, 1); got != image.Rect(256, 256, 300, 300) {
		t.Fatalf("corner tile = %v, want a 44x44 corner", got)
	}

	if b := g.at(1, 1).Bounds(); b != image.Rect(0, 0, 44, 44) {
		t.Fatalf("corner image = %v, want 44x44", b)
	}

	// A tile beside the edge keeps the full tile size: only the far edges
	// clip.
	full := newTileGrid(600, 600)
	if got := full.rectAt(0, 0); got != image.Rect(0, 0, 256, 256) || full.at(1, 1).Bounds() != image.Rect(0, 0, 256, 256) {
		t.Fatalf("interior tile = %v, want a full tile", got)
	}
}
