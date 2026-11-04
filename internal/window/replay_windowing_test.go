package window

import (
	"context"
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

// windowingScreen is a dirtyScreen that observes scroll offsets, the way a
// row-windowed page does.
type windowingScreen struct {
	*dirtyScreen
	ox, oy    int
	windowing bool
}

func (w *windowingScreen) SetScrollOffset(x, y int) { w.ox, w.oy = x, y }
func (w *windowingScreen) ScrollOffset() (int, int) { return w.ox, w.oy }
func (w *windowingScreen) Windowing() bool          { return w.windowing }

func newWindowingShell(inner *dirtyScreen, on bool) *shell {
	app := &windowingScreen{dirtyScreen: inner, windowing: on}

	return &shell{
		app: app, ctx: context.Background(),
		display: inner.display, seq: inner.gen,
		screenW: inner.width, screenH: inner.height,
	}
}

func windowingDisplay(w, h int) *dirtyScreen {
	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 800, height: 600}}
	screen.display = partialDisplay(w, h)
	screen.gen = 1

	return screen
}

// TestWindowingRoutesToViewport checks that a page with the windowing opt-in
// repaints through the viewport-sized buffer even when its content fits the
// tile budget, while the same content without the opt-in keeps the tile grid.
func TestWindowingRoutesToViewport(t *testing.T) {
	t.Parallel()

	plain := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	plain.display = partialDisplay(1920, 2200)
	plain.gen = 1
	s := newDirtyShell(plain)
	dst := partialBuf(1920, 1080)
	defer dst.Deallocate()
	s.drawReplayPartial(dst)
	if s.viewport.buf != nil {
		t.Fatal("an ordinary page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	w := newWindowingShell(inner, true)
	wdst := partialBuf(1920, 1080)
	defer wdst.Deallocate()
	w.drawReplayPartial(wdst)
	if w.viewport.buf == nil {
		t.Fatal("a windowing page kept the content grid")
	}
	if w.partial.tiles != nil {
		t.Fatal("the content grid was retained")
	}
	if got := w.viewport.buf.Bounds().Size(); got != image.Pt(1920, 1080+viewportSides*viewportOverscan) {
		t.Fatalf("viewport buffer = %v, want the window plus overscan", got)
	}
}

// TestWindowingOptOutKeepsTilePath checks that a page carrying the scroll
// observer role but opted out takes exactly the old path.
func TestWindowingOptOutKeepsTilePath(t *testing.T) {
	t.Parallel()

	inner := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	inner.display = partialDisplay(1920, 2200)
	inner.gen = 1
	s := newWindowingShell(inner, false)
	dst := partialBuf(1920, 1080)
	defer dst.Deallocate()
	s.drawReplayPartial(dst)
	if s.viewport.buf != nil {
		t.Fatal("an opted-out page took the viewport path")
	}
	if s.partial.tiles == nil {
		t.Fatal("no content grid")
	}
}

// TestWindowingKeepsScrollRange checks that the opt-in changes only the
// repaint working set: the scrollable content size, which carries the full
// spacer height, is identical with and without windowing.
func TestWindowingKeepsScrollRange(t *testing.T) {
	t.Parallel()

	plain := windowingDisplay(800, 40200)
	plain.boxes = []layout.Box{{W: 800, H: 40200}}
	s := newDirtyShell(plain)
	plainW, plainH := s.contentSize()

	inner := windowingDisplay(800, 40200)
	inner.boxes = []layout.Box{{W: 800, H: 40200}}
	w := newWindowingShell(inner, true)
	winW, winH := w.contentSize()

	if plainW != winW || plainH != winH {
		t.Fatalf("content = %dx%d windowing, %dx%d ordinary", winW, winH, plainW, plainH)
	}
	if winW != 800 || winH != 40200 {
		t.Fatalf("content = %dx%d, want the full spacer height", winW, winH)
	}
}

// TestWindowingWarmupBoundedToViewport checks that the cache-warming sweep
// for a windowed page covers the viewport plus overscan instead of the full
// spacer height.
func TestWindowingWarmupBoundedToViewport(t *testing.T) {
	t.Parallel()

	var warmup viewportWarmup
	content, view := image.Pt(800, 40200), image.Pt(800, 600)
	warmup.take(content, view, image.Point{}, 1, true)
	if warmup.area.Empty() {
		t.Fatal("no windowing warmup area")
	}
	if want := view.Y + viewportSides*viewportOverscan; warmup.area.Dy() != want {
		t.Fatalf("windowing warmup height = %d, want %d", warmup.area.Dy(), want)
	}
	if warmup.area.Dx() != content.X || warmup.area.Min.X != 0 {
		t.Fatalf("windowing warmup width = %v, want the full content width", warmup.area)
	}
	if r := warmup.take(content, view, image.Point{}, 1, true); r.Empty() || r.Dy() > prewarmBandHeight {
		t.Fatalf("windowing warmup band = %v, want one capped band", r)
	}

	scroll := image.Pt(0, 30000)
	warmup.take(content, view, scroll, 1, true)
	wanted := image.Rectangle{Min: scroll, Max: scroll.Add(view)}
	if !wanted.In(warmup.area) {
		t.Fatalf("warmup did not follow the viewport: area=%v", warmup.area)
	}
	if warmup.area.Dy() > view.Y+viewportSides*viewportOverscan {
		t.Fatalf("followed warmup height = %d, want at most the viewport plus overscan", warmup.area.Dy())
	}
}

// TestWindowingWarmupStaysViewportSized drives one viewport frame for a
// compare-like page and checks the sweep the frame scheduled, while the same
// page without the opt-in keeps the old twelve-megapixel neighbourhood.
func TestWindowingWarmupStaysViewportSized(t *testing.T) {
	t.Parallel()

	inner := windowingDisplay(800, 40200)
	w := newWindowingShell(inner, true)
	dst := partialBuf(800, 600)
	defer dst.Deallocate()
	w.drawReplayPartial(dst)
	if got := w.viewport.warmup.area.Dy(); got != 600+viewportSides*viewportOverscan {
		t.Fatalf("windowing sweep height = %d, want the viewport plus overscan", got)
	}

	plain := windowingDisplay(800, 40200)
	s := newDirtyShell(plain)
	sdst := partialBuf(800, 600)
	defer sdst.Deallocate()
	s.drawReplayPartial(sdst)
	if got := s.viewport.warmup.area.Dy(); got != prewarmMaxPixels/800 {
		t.Fatalf("ordinary sweep height = %d, want the old bounded neighbourhood", got)
	}
}
