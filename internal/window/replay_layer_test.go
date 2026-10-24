package window

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/replay"
	"github.com/hajimehoshi/ebiten/v2"
)

// layerDirtyScreen is a dirtyScreen that also names one promoted element.
type layerDirtyScreen struct {
	*dirtyScreen
	specs []replay.LayerSpec
}

func (s *layerDirtyScreen) LayerSpecs(*layout.Display) []replay.LayerSpec {
	return s.specs
}

// windowLayerDisplay is a 100x100 page with a two-op card at (10,10).
func windowLayerDisplay() *layout.Display {
	const pt = 72.0 / 96.0

	return &layout.Display{
		Width: 100, Height: 100, PointsPerPixel: pt, PixelPerPoint: 1,
		Boxes: []layout.Box{
			{ID: "page", X: 0, Y: 0, W: 100, H: 100},
			{ID: "card", X: 10, Y: 10, W: 20, H: 20, ParentIndex: 0},
		},
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 100 * pt, H: 100 * pt, R: 1, G: 1, B: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10 * pt, Y: 10 * pt, W: 20 * pt, H: 10 * pt, R: 1, Alpha: 1},
			{Kind: layout.DisplayOpFillRect, X: 10 * pt, Y: 20 * pt, W: 20 * pt, H: 10 * pt, B: 1, Alpha: 1},
		},
		Order:    []int{0, 1, 2},
		OpOwners: []int{-1, 1, 1},
	}
}

func windowLayerMatrix(tx float64) layout.Matrix2D {
	return layout.Matrix2D{A: 1, D: 1, E: tx}
}

// TestPartialFrameAddsLayerDamage pins the window seam: a clean page frame
// with a moving retained layer must still repaint the layer's old ∪ new pose,
// and dropping the layer must repaint the pose it leaves behind.
func TestPartialFrameAddsLayerDamage(t *testing.T) {
	d := windowLayerDisplay()
	app := &layerDirtyScreen{
		dirtyScreen: &dirtyScreen{
			fakeScreen: &fakeScreen{width: 100, height: 100},
			display:    d,
			gen:        1,
			ok:         true,
		},
	}
	shell := newDirtyShell(app.dirtyScreen)
	shell.app = app
	shell.display = d
	dst := ebiten.NewImage(100, 100)
	defer dst.Deallocate()

	// The first frame builds the tile grid. Promotion then adds its own
	// damage to an otherwise clean page frame.
	shell.drawReplayPartial(dst)
	app.specs = []replay.LayerSpec{{ElementID: "card", BoxIndex: 1, Matrix: windowLayerMatrix(2), Alpha: 1}}
	shell.drawReplayPartial(dst)
	if !shell.renderer.LayersActive(d) {
		t.Fatal("the window did not activate the promoted layer")
	}
	if shell.partial.mode != repaintRect || len(shell.partial.last) == 0 {
		t.Fatalf("promotion mode = %v, last = %v", shell.partial.mode, shell.partial.last)
	}

	app.specs[0].Matrix = windowLayerMatrix(30)
	shell.drawReplayPartial(dst)
	if shell.partial.mode != repaintRect {
		t.Fatalf("move mode = %v, want a rect repaint", shell.partial.mode)
	}
	if !covers(shell.partial.last, image.Rect(10, 10, 30, 30)) {
		t.Fatalf("move damage %v does not cover the old pose", shell.partial.last)
	}
	if !covers(shell.partial.last, image.Rect(40, 10, 60, 30)) {
		t.Fatalf("move damage %v does not cover the new pose", shell.partial.last)
	}

	app.specs = nil
	shell.drawReplayPartial(dst)
	if shell.renderer.LayersActive(d) {
		t.Fatal("a dropped spec left the layer active")
	}
	if shell.partial.mode != repaintRect {
		t.Fatalf("demotion mode = %v, want a rect repaint", shell.partial.mode)
	}
}

// covers reports whether one of rects contains want.
func covers(rects []image.Rectangle, want image.Rectangle) bool {
	for _, rect := range rects {
		if want.In(rect) {
			return true
		}
	}

	return false
}
