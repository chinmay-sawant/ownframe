package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

type fakeSlicerScreen struct {
	*fakeScreen
	sliceReads int
}

func (s *fakeSlicerScreen) BoxesSlice() []layout.Box {
	s.sliceReads++
	return s.boxes
}

var _ host.BoxSlicer = (*fakeSlicerScreen)(nil)

func TestShellBoxesPreference(t *testing.T) {
	displayBoxes := []layout.Box{{ID: "disp", X: 10, Y: 10, W: 50, H: 50}}
	slicerBoxes := []layout.Box{{ID: "slicer", X: 0, Y: 0, W: 100, H: 100}}
	fallbackBoxes := []layout.Box{{ID: "fallback", X: 0, Y: 0, W: 200, H: 200}}

	slicerApp := &fakeSlicerScreen{
		fakeScreen: &fakeScreen{boxes: slicerBoxes},
	}
	s := &shell{
		app: slicerApp,
		display: &layout.Display{
			Boxes: displayBoxes,
		},
	}

	// 1. When display has boxes, s.boxes() returns s.display.Boxes directly.
	got := s.boxes()
	if len(got) != 1 || got[0].ID != "disp" || &got[0] != &displayBoxes[0] {
		t.Fatalf("expected display boxes, got: %+v", got)
	}
	if slicerApp.sliceReads != 0 {
		t.Fatalf("sliceReads = %d, want 0 when display.Boxes is present", slicerApp.sliceReads)
	}

	// 2. When display is nil, s.boxes() calls host.BoxSlicer.
	s.display = nil
	got = s.boxes()
	if len(got) != 1 || got[0].ID != "slicer" || &got[0] != &slicerBoxes[0] {
		t.Fatalf("expected slicer boxes, got: %+v", got)
	}
	if slicerApp.sliceReads != 1 {
		t.Fatalf("sliceReads = %d, want 1", slicerApp.sliceReads)
	}

	// 3. Fallback when app doesn't implement host.BoxSlicer
	fallbackApp := &fakeScreen{boxes: fallbackBoxes}
	s.app = fallbackApp
	got = s.boxes()
	if len(got) != 1 || got[0].ID != "fallback" {
		t.Fatalf("expected fallback boxes, got: %+v", got)
	}

	// 4. boxAt hit-testing
	box, ok := s.boxAt(50, 50)
	if !ok || box.ID != "fallback" {
		t.Fatalf("boxAt(50, 50) = %+v, %v, want fallback", box, ok)
	}
	_, ok = s.boxAt(500, 500)
	if ok {
		t.Fatal("boxAt(500, 500) want false")
	}

	// 5. Zero allocations for s.boxes() and pointer-move hit-testing
	s.app = slicerApp
	s.display = &layout.Display{Boxes: displayBoxes}

	allocs := testing.AllocsPerRun(10, func() {
		_ = s.boxes()
		_, _ = s.boxAt(20, 20)
		s.devHover(20, 20)
	})
	if allocs != 0 {
		t.Fatalf("shell pointer-move operations allocated %f times, want 0", allocs)
	}
}

func BenchmarkPointerMoveBoxHit(b *testing.B) {
	boxes := make([]layout.Box, 100)
	for i := range boxes {
		boxes[i] = layout.Box{
			ID: "box",
			X:  float64(i * 10),
			Y:  float64(i * 10),
			W:  50,
			H:  50,
		}
	}
	s := &shell{
		display: &layout.Display{Boxes: boxes},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x := float64((i % 100) * 10)
		s.devHover(x, x)
	}
}

