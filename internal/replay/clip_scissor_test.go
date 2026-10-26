package replay

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// TestScissorOverflowClipsTransformedPaint pins the composite-time clip: an
// op with a recorded overflow box draws only inside it, and a clip that
// erases the op skips the draw. Ops without a clip keep the full dst.
func TestScissorOverflowClipsTransformedPaint(t *testing.T) {
	dst := ebiten.NewImage(64, 64)

	op := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 48, H: 48}
	op.SetOverflowClip(6, 6, 24, 24)

	sub, paint := scissorOverflow(dst, &op, 0, 0)
	if !paint {
		t.Fatal("clipped op skipped")
	}

	if got := sub.Bounds(); got != image.Rect(8, 8, 40, 40) {
		t.Fatalf("scissor = %v, want (8,8)-(40,40)", got)
	}

	outside := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8}
	outside.SetOverflowClip(48, 48, 8, 8)

	if _, paint := scissorOverflow(dst, &outside, 0, 0); paint {
		t.Fatal("fully clipped op not skipped")
	}

	plain := layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: 0, Y: 0, W: 8, H: 8}
	if _, _, _, _, ok := plain.OverflowClip(); ok {
		t.Fatal("plain op reports a clip")
	}

	if sub, paint := scissorOverflow(dst, &plain, 0, 0); !paint || sub.Bounds() != dst.Bounds() {
		t.Fatalf("plain op dst changed: paint=%v bounds=%v", paint, sub.Bounds())
	}
}
