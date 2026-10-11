package window

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestTransparentWindowDoesNotAddBackground(t *testing.T) {
	s := &shell{transparent: true, display: &layout.Display{
		Width: 240, PointsPerPixel: 0.75,
		Ops: []layout.DisplayOp{{Kind: layout.DisplayOpFillRect, W: 180, H: 165, Alpha: 1}},
	}}
	if _, _, _, alpha := s.pageBackground().RGBA(); alpha != 0 {
		t.Fatalf("transparent window background alpha = %d", alpha)
	}

	s.transparent = false
	if _, _, _, alpha := s.pageBackground().RGBA(); alpha != 65535 {
		t.Fatalf("normal window background alpha = %d", alpha)
	}
}
