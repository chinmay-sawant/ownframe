package window

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image"
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

// ensureTiles keeps the tile grid the size of the content. A content size
// change disposes the old grid and allocates the new one; a zero size leaves
// the window without a grid, as the old single buffer did.
func (s *shell) ensureTiles(width, height int) {
	if s.partial.gridCurrent(width, height) {
		return
	}

	s.dropTiles()

	if width <= 0 || height <= 0 {
		return
	}

	s.partial.tiles = newTileGrid(width, height)
	s.partial.buf = s.partial.tiles.handle()
}

// dropTiles releases the grid and the retained-content handle that stands
// for it.
func (s *shell) dropTiles() {
	if s.partial.tiles != nil {
		s.partial.tiles.dispose()
		s.partial.tiles = nil
	}

	s.partial.buf = nil
}

// dropRetainedTiles drops the grid when a caller dropped the retained-content
// handle. The handle is what a reader asks when it wants to know whether
// retained pixels exist, so losing it invalidates the frame's tiles.
func (s *shell) dropRetainedTiles() {
	if s.partial.tiles != nil && s.partial.buf == nil {
		s.dropTiles()
	}
}

// tileRange returns the half-open tile span that rect covers, clamped to the
// content. A rect that misses the content gives an empty span, and a rect
// that reaches past the content is cut to it.
func (g *tileGrid) tileRange(rect image.Rectangle) (col0, row0, col1, row1 int) {
	rect = rect.Intersect(image.Rect(0, 0, g.width, g.height))
	if rect.Empty() {
		return 0, 0, 0, 0
	}

	return rect.Min.X / tileSize, rect.Min.Y / tileSize,
		(rect.Max.X-1)/tileSize + 1, (rect.Max.Y-1)/tileSize + 1
}

// content is the whole grid in page pixels.
func (g *tileGrid) content() image.Rectangle {
	return image.Rect(0, 0, g.width, g.height)
}

// handle is the retained-content presence handle: a zero-area sub-image of
// the grid's first tile. partialState.buf carries it so a reader that asks
// whether retained pixels exist, such as the window's idle pacing, sees a
// buffer, and a blit from it copies nothing. The pixels live in the tiles.
func (g *tileGrid) handle() *ebiten.Image {
	if g == nil || len(g.images) == 0 || g.images[0] == nil {
		return nil
	}

	img, _ := g.images[0].SubImage(image.Rectangle{}).(*ebiten.Image)

	return img
}

// dispose releases every tile.
func (g *tileGrid) dispose() {
	for _, img := range g.images {
		if img != nil {
			img.Deallocate()
		}
	}
}
