package window

import (
	"image"
	"testing"
)

// TestRebuildCoversEveryTile checks the first frame: the grid is built for
// the content and every tile is painted.
func TestRebuildCoversEveryTile(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)

	s.drawReplayPartial(partialBuf(600, 300))

	if s.partial.mode != repaintBuffer {
		t.Fatalf("mode = %v, want a rebuild", s.partial.mode)
	}

	g := s.partial.tiles
	if g == nil || g.cols != 3 || g.rows != 2 {
		t.Fatalf("grid = %+v, want 3x2 tiles", g)
	}

	if s.partial.paints != g.cols*g.rows {
		t.Fatalf("painted %d tiles, want %d", s.partial.paints, g.cols*g.rows)
	}

	if s.partial.buf == nil || !s.partial.buf.Bounds().Empty() {
		t.Fatal("the retained-content handle should be an empty marker")
	}

	if len(s.partial.last) != 1 || s.partial.last[0] != image.Rect(0, 0, 600, 300) {
		t.Fatalf("last = %v, want the full content", s.partial.last)
	}
}

// TestDirtyRectPaintsOnlyItsTiles checks a click repaints the tiles its box
// meets and no others.
func TestDirtyRectPaintsOnlyItsTiles(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 600, height: 300}}
	screen.display = partialDisplay(600, 300)
	screen.gen = 1
	s := newDirtyShell(screen)
	dst := partialBuf(600, 300)

	s.drawReplayPartial(dst)

	screen.click(image.Rect(10, 10, 20, 20))
	s.drawReplayPartial(dst)

	if s.partial.mode != repaintRect || s.partial.paints != 1 {
		t.Fatalf("small rect: mode = %v, paints = %d, want one tile",
			s.partial.mode, s.partial.paints)
	}

	cross := image.Rect(250, 250, 300, 300)
	screen.click(cross)
	s.drawReplayPartial(dst)

	if s.partial.paints != 4 {
		t.Fatalf("across tiles: painted %d tiles, want 4", s.partial.paints)
	}

	if len(s.partial.last) != 1 || s.partial.last[0] != cross {
		t.Fatalf("last = %v, want the click rect", s.partial.last)
	}
}
