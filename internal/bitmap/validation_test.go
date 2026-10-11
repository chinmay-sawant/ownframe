package bitmap_test

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/bitmap"
)

func TestBitmapRejectsUnknownPaintAndHugeCanvas(t *testing.T) {
	for _, d := range []*layout.Display{
		{Width: 10, Height: 10, Ops: []layout.DisplayOp{{Kind: layout.DisplayKind(100)}}, Order: []int{0}},
		{Width: 1 << 20, Height: 1 << 20},
	} {
		if _, err := bitmap.Paint(d); err == nil {
			t.Fatal("unsupported painting succeeded")
		}
	}
}
