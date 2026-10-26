package window

import (
	"image"
	"testing"
)

// dockShell is a page with the devtools dock open. A 600 px window gives the
// panel the 220 px minimum, so the page clips at x 380.
func dockShell(width, height int) (*dirtyScreen, *shell) {
	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: width, height: height}}
	screen.display = partialDisplay(width, height)
	screen.gen = 1
	s := newDirtyShell(screen)
	s.dev.on = true

	return screen, s
}

// TestDockClipDropsHiddenDamage checks a dirty rect fully under the dock
// repaints no tile and blits none.
func TestDockClipDropsHiddenDamage(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.paints != 4 || s.partial.paintPx != 380*300 {
		t.Fatalf("rebuild = %d tiles %d px, want the 4 left of the dock",
			s.partial.paints, s.partial.paintPx)
	}

	screen.click(image.Rect(400, 10, 420, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintBlit || s.partial.paints != 0 || s.partial.blits != 0 {
		t.Fatalf("hidden damage: mode = %v, paints = %d, blits = %d, want a bare blit",
			s.partial.mode, s.partial.paints, s.partial.blits)
	}
}

// TestDockClipTrimsPartialTiles checks a tile the dock half covers repaints
// and blits only its visible part.
func TestDockClipTrimsPartialTiles(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.click(image.Rect(250, 10, 420, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(250, 10, 380, 20) {
		t.Fatalf("plan = %v %v, want the visible part", s.partial.mode, s.partial.last)
	}

	if s.partial.paintPx != 1300 || s.partial.blits != 2 || s.partial.blitPx != 1300 {
		t.Fatalf("paint = %d px, blits = %d (%d px), want 1300 px in two tile blits",
			s.partial.paintPx, s.partial.blits, s.partial.blitPx)
	}
}
