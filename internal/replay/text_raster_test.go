package replay

import (
	"image"
	"math"
	"testing"
)

func phaseOf(x float64) float64 {
	return textPhase(x - math.Floor(x))
}

// TestTextPhaseAbsorbsFloatNoise locks the sub-pixel grid. Adding whole
// pixels to an origin rounds its low bits, and a key that treated that as a
// new offset would rebake a scrolled list on every step.
func TestTextPhaseAbsorbsFloatNoise(t *testing.T) {
	t.Parallel()

	base := 12.25 * pxPerPt
	for shift := 1; shift <= 64; shift++ {
		if phaseOf(base) != phaseOf(base+float64(shift)) {
			t.Fatalf("phase moved by %d whole pixels", shift)
		}
	}
}

// TestTextRasterLayoutPadsAndCaps checks the image a bake would allocate: the
// ink box grows by the pad on every side, and a run past a cap is declined.
func TestTextRasterLayoutPadsAndCaps(t *testing.T) {
	t.Parallel()

	size, dx, dy, ok := textRasterLayout(image.Rect(20, 30, 120, 50))
	if !ok || dx != 16 || dy != 26 || size != image.Pt(108, 28) {
		t.Fatalf("layout = %v %d %d %v", size, dx, dy, ok)
	}

	if _, _, _, ok := textRasterLayout(image.Rect(0, 0, textRasterMaxSide+1, 20)); ok {
		t.Fatal("an oversize wide run was cacheable")
	}
	if _, _, _, ok := textRasterLayout(image.Rect(0, 0, 20, textRasterMaxSide+1)); ok {
		t.Fatal("an oversize tall run was cacheable")
	}
}
