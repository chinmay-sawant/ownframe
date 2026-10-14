package window

import (
	"image"
	"testing"
)

func TestIdleCleanFrameProbe(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.idleCleanFrame() {
		t.Fatal("a buffer rebuild must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.idleCleanFrame() {
		t.Fatal("a blit-only frame must read clean")
	}

	screen.click(image.Rect(10, 10, 30, 30))
	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.idleCleanFrame() {
		t.Fatal("a repainted frame must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	s.dev.on = true

	if s.idleCleanFrame() {
		t.Fatal("open devtools must read dirty")
	}

	s.dev.on = false
	s.fallback = true

	if s.idleCleanFrame() {
		t.Fatal("a bitmap fallback must read dirty")
	}
}

func TestIdleCleanFrameTicking(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ticking = true
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	s.frameOptIn = true
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.idleCleanFrame() {
		t.Fatal("a ticking page that opted in must read clean")
	}

	s.frameOptIn = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.idleCleanFrame() {
		t.Fatal("a ticking page without the opt-in replays in full")
	}
}
