package window

import (
	"context"
	"image"
	"testing"
)

// lockedScreen is a page that fills the window, so the fitted path runs.
type lockedScreen struct {
	*dirtyScreen
}

func (l *lockedScreen) ViewLocked() bool { return true }

// TestScaledPathKeepsLargePage checks that a large page draws through the
// scaled buffer, so the zoom keeps its scale.
func TestScaledPathKeepsLargePage(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 2200)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(64, 64)

	s.drawReplayScaled(dst)

	if s.replayBuf == nil {
		t.Fatal("a large page fell back to a direct replay")
	}
	if got := s.replayBuf.Bounds().Size(); got != image.Pt(1920, 2200) {
		t.Fatalf("buffer = %v", got)
	}
	s.replayBuf.Dispose()
}

// TestFittedPathKeepsLargePage checks that a large page draws through the
// fitted buffer, so the fit keeps its scale.
func TestFittedPathKeepsLargePage(t *testing.T) {
	t.Parallel()

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	app := &lockedScreen{dirtyScreen: inner}
	s := &shell{
		app: app, ctx: context.Background(), display: inner.display,
		seq: inner.gen, screenW: 1920, screenH: 1080,
	}
	dst := partialBuf(64, 64)

	s.drawFitted(dst)

	if s.fitBuf == nil {
		t.Fatal("a large page fell back to a direct replay")
	}
	if got := s.fitBuf.Bounds().Size(); got != image.Pt(1920, 2200) {
		t.Fatalf("buffer = %v", got)
	}
	s.fitBuf.Dispose()
}
