package window

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

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
