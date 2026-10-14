package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// tileSize is the side of one replay tile in CSS pixels. The content is cut
// into this grid, so a dirty rect repaints only its tiles and a frame blits
// only the damaged ones. Edge tiles are clipped to the content: the grid's
// total pixels equal the content area, the cost of the single content-sized
// buffer it replaces, and no tile lies outside the content.
const tileSize = 256

// tileGrid is the content as a row-major grid of tile images.
type tileGrid struct {
	images []*ebiten.Image
	cols   int
	rows   int
	width  int
	height int
}

// newTileGrid allocates one image per tile, clipped to the content. A size
// with no pixels has no grid.
func newTileGrid(width, height int) *tileGrid {
	if width <= 0 || height <= 0 {
		return nil
	}

	g := &tileGrid{
		cols: gridTiles(width), rows: gridTiles(height),
		width: width, height: height,
	}
	g.images = make([]*ebiten.Image, 0, g.cols*g.rows)

	for row := 0; row < g.rows; row++ {
		for col := 0; col < g.cols; col++ {
			r := g.rectAt(col, row)
			g.images = append(g.images, ebiten.NewImage(r.Dx(), r.Dy()))
		}
	}

	return g
}

// gridTiles is the number of tiles needed to cover n pixels.
func gridTiles(n int) int {
	return (n + tileSize - 1) / tileSize
}

// rectAt returns a tile's rect in page pixels, clipped at the content edge.
func (g *tileGrid) rectAt(col, row int) image.Rectangle {
	x, y := col*tileSize, row*tileSize

	return image.Rect(x, y, min(x+tileSize, g.width), min(y+tileSize, g.height))
}

// at returns the tile image at col, row.
func (g *tileGrid) at(col, row int) *ebiten.Image {
	return g.images[row*g.cols+col]
}
