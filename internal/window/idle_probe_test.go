package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
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
	screen.width, screen.height = s.viewSize()

	if !s.idleCleanFrame() {
		t.Fatal("a stable page with open devtools must still idle")
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

	if !s.idleCleanFrame() {
		t.Fatal("a ticking page with no damage must still idle")
	}
}

func TestFilteredFrameIdleProbeTracksBackdropChanges(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.display.Ops[1].SetBlendGroup(&layout.DisplayGroup{BackdropFilter: "blur(12px)"})
	screen.gen = 1
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("first filtered full replay must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("stable filtered frame must be eligible to idle")
	}

	screen.click(image.Rect(20, 20, 40, 40))
	s.frameOptIn = true
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a filtered frame changed by the ticker must read dirty")
	}

	screen.ok = false
	s.frameOptIn = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("filtered frame should become idle eligible after it stabilizes")
	}
}
