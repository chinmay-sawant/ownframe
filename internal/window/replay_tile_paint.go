package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintTiles repaints the tiles that rect covers, trimmed to the page the
// devtools dock leaves visible. Each tile's part of the rect is cleared to
// the page background, then the operations that meet it are replayed with
// the tile origin subtracted, so page coordinates land on the tile. Fixed
// and viewport-pinned operations stay off the tiles, as they stayed off the
// single buffer.
func (s *shell) paintTiles(d *layout.Display, rect image.Rectangle) {
	g := s.partial.tiles
	if g == nil {
		return
	}

	rect = clipDockX(rect, s.dockClipX(g.width))
	col0, row0, col1, row1 := g.tileRange(rect)

	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			s.paintTile(d, g, col, row, rect)
		}
	}
}

// paintTile clears and replays one tile's part of rect.
func (s *shell) paintTile(d *layout.Display, g *tileGrid, col, row int, rect image.Rectangle) {
	tile := g.rectAt(col, row)
	part := rect.Intersect(tile)
	if part.Empty() {
		return
	}

	sub, ok := g.at(col, row).SubImage(part.Sub(tile.Min)).(*ebiten.Image)
	if !ok {
		return
	}

	s.partial.paints++
	s.partial.paintPx += part.Dx() * part.Dy()
	sub.Fill(s.pageBackground())
	replay.DrawVisibleUnfixedZ(sub, d, -float64(tile.Min.X), -float64(tile.Min.Y), s.viewportPinZ())
}
