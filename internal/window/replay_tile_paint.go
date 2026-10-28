package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintTiles repaints the tiles that rect covers, trimmed to the page the
// devtools dock leaves visible and to the viewport the window shows. A tile
// fully outside the viewport keeps its old pixels out of the paint: damage
// marks it dirty and the blit repaints it when a scroll brings it on screen.
// Fixed and viewport-pinned operations stay off the tiles, as they stayed off
// the single buffer.
func (s *shell) paintTiles(d *layout.Display, rect image.Rectangle) {
	g := s.partial.tiles
	if g == nil {
		return
	}

	rect = clipDockX(rect, s.dockClipX(g.width))
	vis := clipDockX(s.visibleRect(), s.dockClipX(g.width))
	col0, row0, col1, row1 := g.tileRange(rect)

	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			if !g.rectAt(col, row).Overlaps(vis) {
				g.markDirty(col, row)

				continue
			}

			s.paintTile(d, g, col, row, rect)
		}
	}
}

// paintTile clears and replays one tile's part of rect, and marks the tile
// clean: the tile now holds current pixels for its whole area.
func (s *shell) paintTile(d *layout.Display, g *tileGrid, col, row int, rect image.Rectangle) {
	tile := g.rectAt(col, row)
	part := rect.Intersect(tile)
	if part.Empty() {
		return
	}

	tileImg := g.at(col, row)
	subRect := part.Sub(tile.Min)
	var sub *ebiten.Image
	if subRect == tileImg.Bounds() {
		sub = tileImg
	} else {
		var ok bool
		sub, ok = tileImg.SubImage(subRect).(*ebiten.Image)
		if !ok {
			return
		}
	}

	s.partial.paints++
	s.partial.paintPx += part.Dx() * part.Dy()
	g.markClean(col, row)
	sub.Fill(s.pageBackground())
	if s.renderer.LayersActive(d) {
		s.renderer.DrawVisibleUnfixedZLayers(sub, d, -float64(tile.Min.X), -float64(tile.Min.Y), s.viewportPinZ())

		return
	}
	s.renderer.DrawVisibleUnfixedZ(sub, d, -float64(tile.Min.X), -float64(tile.Min.Y), s.viewportPinZ())
}
