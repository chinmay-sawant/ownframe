package window

import (
	"image"
	"testing"
)

// TestDockCloseRepaintsHiddenStrip checks closing the dock repaints the page
// strip the panel hid, even though the page reports no damage.
func TestDockCloseRepaintsHiddenStrip(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.click(image.Rect(300, 10, 320, 20))
	s.drawReplayPartial(dst)

	s.dev.on = false
	screen.rects, screen.ok = nil, false
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || len(s.partial.last) != 1 ||
		s.partial.last[0] != image.Rect(380, 0, 600, 300) {
		t.Fatalf("close = %v %v, want the strip the dock hid", s.partial.mode, s.partial.last)
	}

	if s.partial.paintPx != 220*300 || s.partial.paints != 4 {
		t.Fatalf("close painted %d px in %d tiles, want the 220 px strip",
			s.partial.paintPx, s.partial.paints)
	}
}

// TestDockClipIgnoredWhenClosed checks a closed dock clips nothing: damage
// where the panel would sit repaints in full.
func TestDockClipIgnoredWhenClosed(t *testing.T) {
	t.Parallel()

	screen, s := dockShell(600, 300)
	s.dev.on = false
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.partial.paints != 6 || s.partial.paintPx != 600*300 {
		t.Fatalf("rebuild = %d tiles %d px, want the whole content",
			s.partial.paints, s.partial.paintPx)
	}

	screen.click(image.Rect(300, 10, 320, 20))
	s.drawReplayPartial(dst)

	if s.partial.paintPx != 20*10 || s.partial.blits != 1 {
		t.Fatalf("closed damage = %d px in %d blits, want 200 px in one",
			s.partial.paintPx, s.partial.blits)
	}
}
