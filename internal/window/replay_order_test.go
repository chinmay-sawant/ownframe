package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestScrollBufferRespectsFixedPaintOrder(t *testing.T) {
	s := &shell{app: &fakeScreen{}, display: &layout.Display{
		Ops:   []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, Fixed: true}, {Kind: layout.DisplayOpFillRect}},
		Order: []int{0, 1},
	}}
	if s.bufferOrderSafe() {
		t.Fatal("scroll buffer would cover higher content with lower fixed layer")
	}
	s.display = &layout.Display{Ops: s.display.Ops, Order: []int{1, 0}}
	if !s.bufferOrderSafe() {
		t.Fatal("top viewport layer can use cached scroll buffer")
	}
}
