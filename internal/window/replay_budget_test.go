package window

import (
	"testing"
)

// TestOversizedIgnoresLargePage checks that only a size no image can hold
// counts as oversized. A large page keeps its scale on the scaled and the
// fitted paths.
func TestOversizedIgnoresLargePage(t *testing.T) {
	t.Parallel()

	if oversized(1920, 2200) {
		t.Fatal("a large page must not count as oversized")
	}
	if oversized(3840, 2160) {
		t.Fatal("a 4K page must not count as oversized")
	}
}

// TestGridTooBigMeasuresTheWindow checks the tile grid budget. It is a
// ratio of the content area to the window area, so a page a few screens
// tall keeps the grid and a page many times the window does not.
func TestGridTooBigMeasuresTheWindow(t *testing.T) {
	t.Parallel()

	s := &shell{screenW: 1920, screenH: 1080}
	if s.gridTooBig(1920, 2200) {
		t.Fatal("a page twice the window must keep the tile grid")
	}
	if !s.gridTooBig(1920, 19200) {
		t.Fatal("a page ten times the window must take the viewport path")
	}

	k := &shell{screenW: 3840, screenH: 2160}
	if k.gridTooBig(3840, 2160) {
		t.Fatal("a full-screen 4K page must keep the tile grid")
	}
	if !k.gridTooBig(3840, 43200) {
		t.Fatal("a 4K page ten times the window must take the viewport path")
	}
}

// TestGridTooBigNeedsAWindow checks that a shell without a window size has
// no budget, so the content grid is measured against nothing.
func TestGridTooBigNeedsAWindow(t *testing.T) {
	t.Parallel()

	s := &shell{}
	if s.gridTooBig(4000, 4000) {
		t.Fatal("a shell without a window size must have no budget")
	}
}
