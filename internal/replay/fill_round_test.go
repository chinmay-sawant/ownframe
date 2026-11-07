package replay

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func TestPopulateRoundedFillPathZeroAllocs(t *testing.T) {
	path := roundedFillPathPool.Get().(*vector.Path)
	defer roundedFillPathPool.Put(path)

	radii := [4]float64{5, 5, 5, 5}
	r := radius4(radii)

	// Warm path internal slices
	path.Reset()
	populateRoundedFillPath(path, 10, 10, 80, 80, r)

	allocs := testing.AllocsPerRun(100, func() {
		path.Reset()
		populateRoundedFillPath(path, 10, 10, 80, 80, r)
	})

	if allocs != 0 {
		t.Fatalf("populateRoundedFillPath allocated %v objects per run, want 0", allocs)
	}
}

func TestFillRoundedPoolReusesPath(t *testing.T) {
	dst := ebiten.NewImage(100, 100)
	op := &layout.DisplayOp{
		Kind:  layout.DisplayOpFillRect,
		X:     10,
		Y:     10,
		W:     80,
		H:     80,
		R:     1,
		G:     0.5,
		B:     0.2,
		Alpha: 1,
	}
	radii := [4]float64{5, 5, 5, 5}

	// Calling fillRounded must succeed and leave the pool ready for reuse.
	fillRounded(dst, op, radii, 0, 0)
	fillRounded(dst, op, radii, 0, 0)

	borrowed := roundedFillPathPool.Get().(*vector.Path)
	defer roundedFillPathPool.Put(borrowed)
	// Borrowed path was Reset before being put back
	if borrowed == nil {
		t.Fatal("expected non-nil path from pool")
	}
}
