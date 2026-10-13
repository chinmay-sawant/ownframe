package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// blitTiles copies the tiles covering rect to dst, each placed where the
// scroll puts it. A rect that misses the content copies nothing.
func (s *shell) blitTiles(dst *ebiten.Image, rect image.Rectangle) {
	g := s.partial.tiles
	if g == nil || dst == nil {
		return
	}

	col0, row0, col1, row1 := g.tileRange(rect)

	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			s.blitTile(dst, g, col, row, rect)
		}
	}
}

// blitTile copies one tile's part of rect to the screen. The copy replaces
// pixels, so a transparent page erases. The screen intersection decides the
// source, so a rect that reaches past the window copies only what shows.
func (s *shell) blitTile(dst *ebiten.Image, g *tileGrid, col, row int, rect image.Rectangle) {
	tile := g.rectAt(col, row)
	part := rect.Intersect(tile)
	if part.Empty() {
		return
	}

	offset := image.Pt(s.scrollX, s.scrollY)
	at := part.Sub(offset).Intersect(dst.Bounds())
	if at.Empty() {
		return
	}

	src, ok := g.at(col, row).SubImage(at.Add(offset).Sub(tile.Min)).(*ebiten.Image)
	if !ok {
		return
	}

	var op ebiten.DrawImageOptions
	op.Blend = ebiten.BlendCopy
	op.GeoM.Translate(float64(at.Min.X), float64(at.Min.Y))
	dst.DrawImage(src, &op)
}
