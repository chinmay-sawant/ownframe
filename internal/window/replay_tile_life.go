package window

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
