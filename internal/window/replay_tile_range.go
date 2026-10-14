package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

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
			img.Dispose()
		}
	}
}
