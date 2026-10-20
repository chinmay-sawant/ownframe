package window

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// blitFrame shows the frame's tiles on the screen. A stale frame, a grid the
// chrome rects outgrow, or a just-rebuilt grid fills the screen and copies
// every tile; a scroll copies the whole content at the new offset; anything
// else copies only the damaged tiles. The retained screen still holds every
// untouched tile, so the screen is complete when this returns.
func (s *shell) blitFrame(dst *ebiten.Image, display *layout.Display, plans []repaintPlan, chrome []image.Rectangle, whole, stale bool) {
	grid := s.partial.tiles
	if stale || grid == nil || !withinTiles(grid, chrome) {
		if !s.showTilesWhole(dst) {
			s.paintFullReplay(dst, display)
		}

		return
	}

	full, rects := blitPlan(plans, chrome)
	if full {
		s.showTilesWhole(dst)

		return
	}

	if whole {
		s.blitTiles(dst, s.visibleRect())
		return
	}

	for _, rect := range rects {
		s.blitTiles(dst, rect)
	}
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
