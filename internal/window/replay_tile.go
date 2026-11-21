package window

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
)

// tileSize is the side of one replay tile in CSS pixels. The content is cut
// into this grid, so a dirty rect repaints only its tiles and a frame blits
// only the damaged ones. Edge tiles are clipped to the content: the grid's
// total pixels equal the content area, the cost of the single content-sized
// buffer it replaces, and no tile lies outside the content.
const tileSize = 256

// tileGrid is the content as a row-major grid of tile images. clean tracks
// which tiles hold current pixels: a tile turns dirty when damage covers it
// while it sits outside the visible viewport, and the blit repaints it on
// demand before copying. Skipping offscreen tiles keeps a full-content
// damage rect at viewport cost; a scroll catches the skipped tiles up.
type tileGrid struct {
	images []*ebiten.Image
	clean  []bool
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
		clean: make([]bool, gridTiles(width)*gridTiles(height)),
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

// cleanAt reports whether the tile holds current pixels. A fresh grid
// starts dirty: nothing is painted until a paint or an on-demand blit
// fills it.
func (g *tileGrid) cleanAt(col, row int) bool {
	return g.clean[row*g.cols+col]
}

// markClean records that the tile was just repainted.
func (g *tileGrid) markClean(col, row int) {
	g.clean[row*g.cols+col] = true
}

// markDirty records that damage covered a tile the paint pass skipped, so a
// later blit repaints it before copying.
func (g *tileGrid) markDirty(col, row int) {
	g.clean[row*g.cols+col] = false
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
	if cap(s.partial.plans) < maxRepaintRects+1 {
		s.partial.plans = make([]repaintPlan, 0, maxRepaintRects+1)
	}
	if cap(s.partial.last) < maxRepaintRects+1 {
		s.partial.last = make([]image.Rectangle, 0, maxRepaintRects+1)
	}
	if cap(s.partial.blitRects) < maxBlitRects+1 {
		s.partial.blitRects = make([]image.Rectangle, 0, maxBlitRects+1)
	}
	s.partial.blitOpts.Blend = ebiten.BlendCopy
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

// blitFrame shows the frame's tiles on the screen. A stale frame, a grid the
// chrome rects outgrow, or a just-rebuilt grid fills the screen and copies
// every tile; a scroll copies the whole content at the new offset; anything
// else copies only the damaged tiles. The retained screen still holds every
// untouched tile, so the screen is complete when this returns.
func (s *shell) blitFrame(dst *ebiten.Image, display *layout.Display, plans []repaintPlan, chrome []image.Rectangle, whole, stale bool) {
	if s.blitFromTiles(dst, display, plans, chrome, whole, stale) {
		s.paintPinnedChrome(dst, display)

		return
	}

	s.paintFullReplay(dst, display)
}

// showTilesWhole fills the screen with the page background and copies every
// tile. The grid is current for the frame, so this replaces a second full
// replay of the display list on a frame that only needs the screen
// repainted. It reports whether the grid could supply the frame. Tiles the
// paint pass left dirty repaint on demand as the copy reaches them.
func (s *shell) showTilesWhole(dst *ebiten.Image, d *layout.Display) bool {
	g := s.partial.tiles
	if g == nil {
		return false
	}

	dst.Fill(s.pageBackground())
	s.blitTiles(dst, d, g.content())

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
// rect that misses the content copies nothing. A tile whose damage arrived
// while it sat offscreen repaints first, so a scroll never copies stale
// pixels.
func (s *shell) blitTiles(dst *ebiten.Image, d *layout.Display, rect image.Rectangle) {
	g := s.partial.tiles
	if g == nil || dst == nil {
		return
	}

	rect = clipDockX(rect, s.dockClipX(g.width))
	col0, row0, col1, row1 := g.tileRange(rect)

	for row := row0; row < row1; row++ {
		for col := col0; col < col1; col++ {
			s.blitTile(dst, d, g, col, row, rect)
		}
	}
}

// blitTile copies one tile's part of rect to the screen. The copy replaces
// pixels, so a transparent page erases. The screen intersection decides the
// source, so a rect that reaches past the window copies only what shows. A
// dirty tile repaints whole before the copy, so a scroll into skipped tiles
// shows current pixels.
func (s *shell) blitTile(dst *ebiten.Image, d *layout.Display, g *tileGrid, col, row int, rect image.Rectangle) {
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

	if !g.cleanAt(col, row) {
		s.paintTile(d, g, col, row, tile)
	}

	tileImg := g.at(col, row)
	srcRect := at.Add(offset).Sub(tile.Min)
	var src *ebiten.Image
	if srcRect == tileImg.Bounds() {
		src = tileImg
	} else {
		var ok bool
		src, ok = tileImg.SubImage(srcRect).(*ebiten.Image)
		if !ok {
			return
		}
	}

	s.partial.blits++
	s.partial.blitPx += at.Dx() * at.Dy()

	s.partial.blitOpts.Blend = ebiten.BlendCopy
	s.partial.blitOpts.GeoM.Reset()
	s.partial.blitOpts.GeoM.Translate(float64(at.Min.X), float64(at.Min.Y))
	dst.DrawImage(src, &s.partial.blitOpts)
}

// blitFromTiles copies the frame from the tile grid. false means the
// tiles cannot supply the screen, so the caller paints a full replay.
func (s *shell) blitFromTiles(dst *ebiten.Image, display *layout.Display, plans []repaintPlan, chrome []image.Rectangle, whole, stale bool) bool {
	grid := s.partial.tiles
	if stale || grid == nil || !withinTiles(grid, chrome) {
		return s.showTilesWhole(dst, display)
	}

	full, rects := blitPlanBuf(s.partial.blitRects[:0], plans, chrome)
	s.partial.blitRects = rects
	if full {
		return s.showTilesWhole(dst, display)
	}

	// Pinned chrome paints over the screen after this blit. Restoring the
	// visible unfixed tiles first keeps translucent fixed layers (such as
	// modal backdrops or shadows) from compositing over their own pixels
	// from the previous retained frame.
	if whole || s.hasViewportOps() {
		vis := s.visibleRect()
		if !vis.In(grid.content()) {
			return s.showTilesWhole(dst, display)
		}

		s.blitTiles(dst, display, vis)

		return true
	}

	for _, rect := range rects {
		s.blitTiles(dst, display, rect)
	}

	return true
}

// dockClipX returns the first page x the devtools dock hides this frame:
// page pixels at or right of it sit under the panel. The panel is placed
// from the live dock width, the same geometry devLayout paints later this
// frame. With the dock off the content width comes back, so nothing clips.
func (s *shell) dockClipX(contentW int) int {
	if !s.dev.on || s.dev.mobile {
		return contentW
	}

	panel := devDockRect(s.screenW, s.screenH, s.devDockWidth())
	x := int(math.Ceil(panel.X)) + s.scrollX
	if x < 0 {
		x = 0
	}

	if x > contentW {
		x = contentW
	}

	return x
}

// clipDockX trims a page rect to the page x the dock leaves visible. The
// dock is full height, so only x clips.
func clipDockX(r image.Rectangle, clipX int) image.Rectangle {
	if r.Max.X > clipX {
		r.Max.X = clipX
	}

	if r.Min.X >= r.Max.X {
		return image.Rectangle{}
	}

	return r
}

// dockStrip is the page strip the previous frame's dock hid and this frame
// shows. Tiles there were not repainted while hidden, so the frame must
// catch them up. Empty unless the visible clip grew.
func (st partialState) dockStrip(clipX, contentH int) image.Rectangle {
	if clipX <= st.clipX || contentH <= 0 {
		return image.Rectangle{}
	}

	return image.Rect(st.clipX, 0, clipX, contentH)
}

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
