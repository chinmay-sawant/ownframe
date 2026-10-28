package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
)

// viewportTaker reports scripted dirty rects for the oversized path.
type viewportTaker struct {
	rects []image.Rectangle
	ok    bool
	takes int
}

func (v *viewportTaker) TakeDirtyRects() ([]image.Rectangle, bool) {
	v.takes++

	return v.rects, v.ok
}

func TestViewportBufferStaysWindowSized(t *testing.T) {
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 80, height: 60}, gen: 1}
	app.display = partialDisplay(80, 37000)
	s := newDirtyShell(app)
	taker := &viewportTaker{}
	dst := partialBuf(80, 60)
	s.partial.tiles = newTileGrid(80, 100)
	s.drawViewport(dst, app.display, taker)
	if got := s.viewport.buf.Bounds(); got != image.Rect(0, 0, 80, 180) {
		t.Fatalf("buffer = %v", got)
	}
	if s.partial.tiles != nil {
		t.Fatal("content grid retained")
	}
	saved := s.viewport.buf
	s.drawViewport(dst, app.display, taker)
	if s.viewport.buf != saved {
		t.Fatal("idle frame reallocated")
	}
	s.scrollY = 2000
	s.drawViewport(dst, app.display, taker)
	if s.viewport.area.Min.Y > 2000 || s.viewport.area.Max.Y < 2060 || s.viewport.buf != saved {
		t.Fatal("scroll lost or replaced the buffer")
	}
	app.gen++
	s.seq = app.gen
	s.drawViewport(dst, app.display, taker)
	if s.viewport.key.generation != app.gen {
		t.Fatal("generation did not invalidate")
	}
	if taker.takes != 4 {
		t.Fatalf("takes = %d, want 4", taker.takes)
	}
	s.disposeViewport()
	if s.viewport.buf != nil || s.viewport.valid {
		t.Fatal("viewport not released")
	}
}

func TestViewportDirtyBounds(t *testing.T) {
	one := image.Rect(10, 20, 30, 40)
	if got := viewportDirtyBounds([]image.Rectangle{one}); got != one {
		t.Fatalf("one rect = %v, want %v", got, one)
	}
	two := []image.Rectangle{one, image.Rect(200, 300, 300, 400)}
	if got := viewportDirtyBounds(two); got != image.Rect(10, 20, 300, 400) {
		t.Fatalf("two rects = %v", got)
	}
	if got := viewportDirtyBounds(nil); !got.Empty() {
		t.Fatalf("empty list = %v, want empty", got)
	}
}

// promotedLayerScreen records the layer ids the window hands back to the page.
type promotedLayerScreen struct {
	*layerDirtyScreen
	promoted []string
}

func (s *promotedLayerScreen) SetPromotedLayers(ids []string) {
	s.promoted = append([]string(nil), ids...)
}

// TestViewportCompositesRetainedLayer pins the background-plus-layer path: a
// tall page bakes the promoted card into a layer, leaves its operations out of
// the cached background, and reports the promoted id back to the page so the
// next sample stops damaging the page under it.
func TestViewportCompositesRetainedLayer(t *testing.T) {
	d := windowLayerDisplay()
	app := &promotedLayerScreen{layerDirtyScreen: &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
			ok:         true,
		},
		specs: []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}},
	}}
	shell := newDirtyShell(app.dirtyScreen)
	shell.app = app
	shell.display = d
	dst := ebiten.NewImage(100, 100)
	defer dst.Deallocate()

	shell.drawViewport(dst, d, &viewportTaker{})

	if shell.viewport.buf == nil {
		t.Fatal("the viewport path kept no background buffer")
	}
	if !shell.renderer.LayersActive(d) {
		t.Fatal("the card was not promoted")
	}
	if len(app.promoted) != 1 || app.promoted[0] != "card" {
		t.Fatalf("promoted ids = %v, want [card]", app.promoted)
	}
}

func TestViewportSpareRetainedAcrossScrollPause(t *testing.T) {
	app := &dirtyScreen{fakeScreen: &fakeScreen{width: 80, height: 60}, gen: 1}
	app.display = partialDisplay(80, 37000)
	s := newDirtyShell(app)
	taker := &viewportTaker{}
	dst := partialBuf(80, 60)
	s.drawViewport(dst, app.display, taker)

	// Scroll crosses overscan, triggering moveViewport which allocates spare.
	s.scrollY = 150
	s.drawViewport(dst, app.display, taker)
	if s.viewport.spare == nil {
		t.Fatal("moveViewport did not allocate spare")
	}
	spare := s.viewport.spare

	// Settle warmup and pause scrolling. Next frame is blitOnly.
	s.viewport.warmup.next.Y = s.viewport.warmup.area.Max.Y
	s.drawViewport(dst, app.display, taker)
	if !s.viewport.blitOnly {
		t.Fatal("stationary frame was not blitOnly")
	}
	if s.viewport.spare == nil || s.viewport.spare != spare {
		t.Fatal("spare was deallocated or replaced on scroll pause")
	}

	// Resuming scroll uses the retained spare.
	s.scrollY = 220
	s.drawViewport(dst, app.display, taker)
	if s.viewport.spare == nil {
		t.Fatal("spare missing after resuming scroll")
	}

	// dropDisplayBuffers explicitly deallocates spare.
	s.dropDisplayBuffers()
	if s.viewport.spare != nil || s.viewport.buf != nil {
		t.Fatal("dropDisplayBuffers did not release viewport buffers")
	}
}
