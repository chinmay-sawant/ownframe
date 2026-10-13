package window

import (
	"image"
	"testing"
)

// dockDamageFrame runs a rebuild and then one damage frame on a page of w x h
// with the dock open or closed. It returns the damage frame's painted area
// and blit count plus the rebuild's painted tiles and area.
func dockDamageFrame(w, h int, dock bool, damage image.Rectangle) (px, blits, tiles, tilePx int) {
	screen, s := dockShell(w, h)
	s.dev.on = dock
	dst := partialBuf(w, h)

	s.drawReplayPartial(dst)
	s.frameCommit()
	tiles, tilePx = s.partial.paints, s.partial.paintPx

	screen.click(damage)
	s.drawReplayPartial(dst)

	return s.partial.paintPx, s.partial.blits, tiles, tilePx
}

// TestDockClipMeasuresTheWork pins the dock clip's savings on a board-sized
// page: a rebuild and a fixed damage rect, closed and open. The numbers go to
// the log for the perf report.
func TestDockClipMeasuresTheWork(t *testing.T) {
	t.Parallel()

	const w, h = 1908, 1080
	damage := image.Rect(1400, 200, 1908, 500)

	cPx, cBlits, cTiles, cTilePx := dockDamageFrame(w, h, false, damage)
	oPx, oBlits, oTiles, oTilePx := dockDamageFrame(w, h, true, damage)

	t.Logf("dock closed: rebuild %d tiles / %d px; damage %d px in %d blits",
		cTiles, cTilePx, cPx, cBlits)
	t.Logf("dock open:   rebuild %d tiles / %d px; damage %d px in %d blits",
		oTiles, oTilePx, oPx, oBlits)

	if cTiles != 40 || cTilePx != w*h || cPx != 508*300 || cBlits != 6 {
		t.Fatalf("closed = rebuild %d tiles / %d px, damage %d px in %d blits",
			cTiles, cTilePx, cPx, cBlits)
	}

	if oTiles != 35 || oTilePx != 1568*h || oPx != 168*300 || oBlits != 4 {
		t.Fatalf("open = rebuild %d tiles / %d px, damage %d px in %d blits",
			oTiles, oTilePx, oPx, oBlits)
	}
}
