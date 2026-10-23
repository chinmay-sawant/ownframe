package window

import (
	"image"
	"testing"
)

func TestPinnedChromeKeepsTilePath(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)

	if !s.hasViewportOps() {
		t.Fatal("fixture has no viewport ops")
	}

	if s.fullReplay() {
		t.Fatal("hasViewportOps forced fullReplay")
	}

	if s.partial.tiles == nil || s.partial.paints == 0 || s.partial.pins == 0 {
		t.Fatal("pinned chrome left the tile path")
	}

	screen.click(image.Rect(10, 80, 90, 120))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || s.partial.tiles == nil || s.partial.pins == 0 {
		t.Fatalf("message dirty: mode = %v pins = %d", s.partial.mode, s.partial.pins)
	}
}

func TestPinnedChromeIdleOnBlit(t *testing.T) {
	t.Parallel()

	screen, s := newPinShell()
	dst := partialBuf(100, 200)

	s.drawReplayPartial(dst)
	s.frameCommit()
	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.hasViewportOps() || s.fullReplay() {
		t.Fatal("idle pin frame must not full-replay")
	}

	if s.partial.pins == 0 {
		t.Fatal("idle blit skipped pinned chrome")
	}

	if !s.idleCleanFrame() {
		t.Fatal("valid tiles with viewport ops must idle")
	}
}
