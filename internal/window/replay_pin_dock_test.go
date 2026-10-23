package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDockClipsPinnedChrome(t *testing.T) {
	t.Parallel()

	d := &layout.Display{
		Width: 600, Height: 300, PointsPerPixel: 1,
		Ops: []layout.DisplayOp{
			pinOp(0, 0, 600, 300, 1, 1, 1, 0),
			pinOp(0, 0, 600, 40, 1, 0, 0, pinLayer),
		},
		Order: []int{0, 1},
	}
	screen := &pinScreen{dirtyScreen: &dirtyScreen{
		fakeScreen: &fakeScreen{width: 600, height: 300},
		display:    d, gen: 1,
	}}
	s := newDirtyShell(screen.dirtyScreen)
	s.app, s.display = screen, d
	s.screenW, s.screenH = 600, 300
	s.dev.on = true
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)

	if s.partial.tiles == nil || !s.hasViewportOps() || s.fullReplay() {
		t.Fatal("docked pin chrome left the tile path")
	}

	clip := s.pinClip(dst)
	viewW, _ := s.viewSize()
	if clip == nil || clip.Bounds().Dx() != viewW || viewW >= 600 {
		t.Fatalf("pin clip = %v viewW = %d, want the dock strip", clip, viewW)
	}

	if s.partial.pins == 0 {
		t.Fatal("docked frame skipped pinned chrome")
	}
}
