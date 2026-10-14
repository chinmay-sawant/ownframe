package window

import (
	"image"
	"testing"
)

// dockShell is a page with the devtools dock open. The default 340 px panel
// starts at page x width-340, so a 600 px page clips at 260.
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

	if s.partial.paints != 4 || s.partial.paintPx != 260*300 {
		t.Fatalf("rebuild = %d tiles %d px, want the 4 left of the dock",
			s.partial.paints, s.partial.paintPx)
	}

	screen.click(image.Rect(300, 10, 320, 20))
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

	screen.click(image.Rect(250, 10, 300, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(250, 10, 260, 20) {
		t.Fatalf("plan = %v %v, want the visible half", s.partial.mode, s.partial.last)
	}

	if s.partial.paintPx != 100 || s.partial.blits != 2 || s.partial.blitPx != 100 {
		t.Fatalf("paint = %d px, blits = %d (%d px), want 100 px in two tile blits",
			s.partial.paintPx, s.partial.blits, s.partial.blitPx)
	}
}
