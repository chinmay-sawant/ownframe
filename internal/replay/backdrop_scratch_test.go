package replay

import (
	"image"
	"testing"
)

// TestScratchSizingClampsAndShrinks pins GPU-LIFE-01: requests past the GPU
// edge limit are refused, and aspect-ratio swings shrink the high-water mark
// after a watch window instead of holding it for the life of the renderer.
func TestScratchSizingClampsAndShrinks(t *testing.T) {
	var s scratchSizing
	const maxEdge = 4096

	if _, ok := s.plan(image.Point{}, maxEdge+1, 10, maxEdge); ok {
		t.Fatal("request beyond MaxImageSize was accepted")
	}
	if _, ok := s.plan(image.Point{}, 0, 10, maxEdge); ok {
		t.Fatal("empty request was accepted")
	}

	size, ok := s.plan(image.Point{}, 200, 2000, maxEdge)
	if !ok || size != image.Pt(200, 2000) {
		t.Fatalf("first plan = %v %v, want 200x2000", size, ok)
	}
	// Aspect swing: a wide sample grows the union (200x2000 + 2000x200).
	size, _ = s.plan(size, 2000, 200, maxEdge)
	if size != image.Pt(2000, 2000) {
		t.Fatalf("after swing size = %v, want 2000x2000", size)
	}

	// A small steady demand shrinks the textures once the window elapses.
	for range 2 * scratchShrinkFrames {
		size, _ = s.plan(size, 100, 100, maxEdge)
	}
	if size != image.Pt(100, 100) {
		t.Fatalf("size after steady small demand = %v, want 100x100", size)
	}

	// Steady demand near the allocation never reallocates.
	for range 3 * scratchShrinkFrames {
		next, _ := s.plan(size, 90, 100, maxEdge)
		if next != size {
			t.Fatalf("steady demand reallocated %v -> %v", size, next)
		}
	}
}

func TestBlurUniformsReuseMap(t *testing.T) {
	var b backdropState
	first := b.blurUniforms(&b.horizontalUniforms, 4, 0, 1, 0, 1, 2, 3, 4, 5)
	second := b.blurUniforms(&b.horizontalUniforms, 8, 0, 1, 0, 1, 2, 3, 4, 5)
	if len(first) != 9 || second["Radius"] != float32(8) {
		t.Fatalf("uniforms = %v", second)
	}
	if allocs := testing.AllocsPerRun(100, func() {
		b.blurUniforms(&b.horizontalUniforms, 8, 0, 1, 0, 1, 2, 3, 4, 5)
	}); allocs != 0 {
		t.Fatalf("blurUniforms allocs/op = %v, want 0", allocs)
	}
}
