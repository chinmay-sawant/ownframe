package ui

// tile is one fetched rectangular range of cells.
type tile struct {
	sheet string
	area  Area
	cells []Cell
	rev   uint64
	used  uint64
}

// tileCache holds a bounded number of fetched tiles, newest first. It never
// materializes a whole sheet.
type tileCache struct {
	tiles    []tile
	maxTiles int
	maxCells int
	cells    int
	clock    uint64
}

func newTileCache() *tileCache {
	return &tileCache{maxTiles: uiTileMax, maxCells: uiTileCells}
}

// get returns one cell from the first tile that holds it.
func (t *tileCache) get(sheet string, r, c int) (Cell, bool) {
	for i := range t.tiles {
		tl := &t.tiles[i]
		if tl.sheet != sheet || c < tl.area.C0 || c > tl.area.C1 || r < tl.area.R0 || r > tl.area.R1 {
			continue
		}

		t.clock++
		tl.used = t.clock
		idx := (r-tl.area.R0)*tl.area.W() + (c - tl.area.C0)
		if idx < 0 || idx >= len(tl.cells) {
			return Cell{}, false
		}

		return tl.cells[idx], true
	}

	return Cell{}, false
}

// put stores a tile, evicting the least recently used tiles until the
// bounds hold.
func (t *tileCache) put(sheet string, a Area, cells []Cell, rev uint64) {
	if a.Empty() || len(cells) < a.Count() || a.Count() > t.maxCells {
		return
	}

	t.dropCovered(sheet, a)

	t.clock++
	t.tiles = append(t.tiles, tile{sheet, a, cells, rev, t.clock})
	t.cells += len(cells)

	for len(t.tiles) > t.maxTiles || t.cells > t.maxCells {
		t.evictOldest()
	}
}

// invalidate drops every tile, for a workbook revision change.
func (t *tileCache) invalidate() {
	t.tiles = t.tiles[:0]
	t.cells = 0
}
