package window

import (
	"context"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestThumbShowsResizeCursor(t *testing.T) {
	t.Parallel()

	app := &fakeScreen{
		width: 200, height: 200, minW: 1, minH: 1,
		boxes: []layout.Box{{W: 500, H: 500}},
	}
	s := newShell(context.Background(), app)

	var got []ebiten.CursorShapeType
	s.setCursor = func(shape ebiten.CursorShapeType) { got = append(got, shape) }

	s.pointerScrollbar(198, 10, false, false)
	s.pointerScrollbar(10, 198, false, false)
	s.pointerScrollbar(100, 100, false, false)

	want := []ebiten.CursorShapeType{ebiten.CursorShapeNSResize, ebiten.CursorShapeEWResize}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("cursor shapes = %v, want %v", got, want)
	}
}
