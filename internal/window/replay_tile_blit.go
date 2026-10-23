package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
)

// blitFrame shows the frame's tiles on the screen. A stale frame, a grid the
// chrome rects outgrow, or a just-rebuilt grid fills the screen and copies
// every tile; a scroll copies the whole content at the new offset; anything
// else copies only the damaged tiles. The retained screen still holds every
// untouched tile, so the screen is complete when this returns.
func (s *shell) blitFrame(dst *ebiten.Image, display *layout.Display, plans []repaintPlan, chrome []image.Rectangle, whole, stale bool) {
	if s.blitFromTiles(dst, plans, chrome, whole, stale) {
		s.paintPinnedChrome(dst, display)

		return
	}

	s.paintFullReplay(dst, display)
}

// showTilesWhole fills the screen with the page background and copies every
// tile. The grid is current for the frame, so this replaces a second full
// replay of the display list on a frame that only needs the screen
// repainted. It reports whether the grid could supply the frame.
func (s *shell) showTilesWhole(dst *ebiten.Image) bool {
	g := s.partial.tiles
	if g == nil {
		return false
	}

	dst.Fill(s.pageBackground())
	s.blitTiles(dst, g.content())

	return true
}

// withinTiles reports that every chrome rect has a source region inside the
// grid. Chrome that reaches past a narrow page cannot be restored, so such a
// frame repaints in full.
func withinTiles(g *tileGrid, rects []image.Rectangle) bool {
	content := g.content()

	for _, rect := range rects {
		if !rect.In(content) {
			return false
		}
	}

	return true
}

// visibleRect is the page area the screen shows at the current scroll. A
// scroll copies only this part of the tile grid, so the copy does not walk
// every tile of a long page.
func (s *shell) visibleRect() image.Rectangle {
	viewW, viewH := s.viewSize()

	return image.Rect(s.scrollX, s.scrollY, s.scrollX+viewW, s.scrollY+viewH)
}

// blitTiles copies the tiles covering rect to dst, each placed where the
// scroll puts it, trimmed to the page the devtools dock leaves visible. A
// rect that misses the content copies nothing.
func (s *shell) blitTiles(dst *ebiten.Image, rect image.Rectangle) {
	g := s.partial.tiles
	if g == nil || dst == nil {
		return
	}

	rect = clipDockX(rect, s.dockClipX(g.width))
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

	s.partial.blits++
	s.partial.blitPx += at.Dx() * at.Dy()

	var op ebiten.DrawImageOptions
	op.Blend = ebiten.BlendCopy
	op.GeoM.Translate(float64(at.Min.X), float64(at.Min.Y))
	dst.DrawImage(src, &op)
}

// blitFromTiles copies the frame from the tile grid. false means the
// tiles cannot supply the screen, so the caller paints a full replay.
func (s *shell) blitFromTiles(dst *ebiten.Image, plans []repaintPlan, chrome []image.Rectangle, whole, stale bool) bool {
	grid := s.partial.tiles
	if stale || grid == nil || !withinTiles(grid, chrome) {
		return s.showTilesWhole(dst)
	}

	full, rects := blitPlan(plans, chrome)
	if full {
		return s.showTilesWhole(dst)
	}

	if whole {
		s.blitTiles(dst, s.visibleRect())

		return true
	}

	for _, rect := range rects {
		s.blitTiles(dst, rect)
	}

	return true
}
