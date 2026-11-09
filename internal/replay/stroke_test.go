package replay

import (
	"image"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestStrokeGeometryZeroRadiiBounds(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    0, Y: 0, W: 30, H: 15, Width: 2,
	}

	path, width, ok := strokeGeometry(op, 3, 7)
	if !ok {
		t.Fatal("zero-radius stroke rejected")
	}

	want := image.Rect(3, 7, 43, 27)
	if got := path.Bounds(); got != want {
		t.Fatalf("bounds = %v, want %v", got, want)
	}

	if wantWidth := float32(2 * pxPerPt); width != wantWidth {
		t.Fatalf("width = %v, want %v", width, wantWidth)
	}
}

func TestStrokeGeometryCircularRadiusPath(t *testing.T) {
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
	}

	path, _, ok := strokeGeometry(op, 0, 0)
	if !ok {
		t.Fatal("circular-radius stroke rejected")
	}

	bounds := path.Bounds()
	if bounds.Empty() {
		t.Fatal("circular-radius path is empty")
	}

	want := image.Rect(13, 6, 54, 27)
	if bounds != want {
		t.Fatalf("bounds = %v, want %v", bounds, want)
	}
}

func TestPopulateRoundedPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	rx := [4]float32{6, 6, 6, 6}
	ry := [4]float32{6, 6, 6, 6}

	// Warm path internal slices
	path.Reset()
	populateRoundedPath(path, 10, 5, 40, 20, rx, ry)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateRoundedPath(path, 10, 5, 40, 20, rx, ry)
	})

	if allocs != 0 {
		t.Fatalf("populateRoundedPath allocated %v objects per run, want 0", allocs)
	}
}

func TestPopulateMaskPathZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	// Warm path internal slices
	path.Reset()
	populateMaskPath(path, 10, 5, 40, 20, 15)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateMaskPath(path, 10, 5, 40, 20, 15)
	})

	if allocs != 0 {
		t.Fatalf("populateMaskPath allocated %v objects per run, want 0", allocs)
	}
}

func TestPopulateStrokeGeometryZeroAllocs(t *testing.T) {
	path := acquireVectorPath()
	defer releaseVectorPath(path)

	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
	}

	// Warm path internal slices
	path.Reset()
	populateStrokeGeometry(path, op, 0, 0)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateStrokeGeometry(path, op, 0, 0)
	})

	if allocs != 0 {
		t.Fatalf("populateStrokeGeometry allocated %v objects per run, want 0", allocs)
	}
}

func TestStrokeDirectPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpStrokeRect,
		X:    10, Y: 5, W: 30, H: 15, Width: 1, Radius: 6,
		R: 1, G: 0, B: 0, Alpha: 1,
	}

	// Calling strokeDirect must succeed and leave the pool ready for reuse.
	strokeDirect(dst, op, 0, 0)
	strokeDirect(dst, op, 0, 0)

	borrowed := acquireVectorPath()
	defer releaseVectorPath(borrowed)
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}

func TestDrawSegmentPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind: layout.DisplayOpLine,
		R:    1, G: 0, B: 0, Alpha: 1,
	}

	// Calling drawSegment must succeed and leave the pool ready for reuse.
	drawSegment(dst, op, 10, 10, 50, 50, 2, 0, 0)
	drawSegment(dst, op, 10, 10, 50, 50, 2, 0, 0)

	borrowed := acquireVectorPath()
	defer releaseVectorPath(borrowed)
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}
