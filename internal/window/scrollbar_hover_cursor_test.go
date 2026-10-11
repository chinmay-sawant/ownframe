package window

import (
	"context"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestThumbHoverKeepsResizeCursor(t *testing.T) {
	t.Parallel()

	app := &cursorScreen{fakeScreen: &fakeScreen{
		width: 200, height: 200, minW: 1, minH: 1,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := NewGame(context.Background(), app).(*shell)

	var got []ebiten.CursorShapeType
	s.setCursor = func(shape ebiten.CursorShapeType) { got = append(got, shape) }

	frameW, frameH := s.frameSize()
	thumbs := []struct {
		x, y int
		want ebiten.CursorShapeType
	}{
		{198, 10, ebiten.CursorShapeNSResize},
		{10, 198, ebiten.CursorShapeEWResize},
	}

	for _, tc := range thumbs {
		got = nil
		s.pointerScrollbar(tc.x, tc.y, false, false)
		px, py := s.contentAt(tc.x, tc.y, frameW, frameH)
		if err := s.hoverCursor(tc.x, tc.y, px, py); err != nil {
			t.Fatal(err)
		}
		if s.cursor != tc.want {
			t.Fatalf("hover at %d,%d cursor = %v, want %v", tc.x, tc.y, s.cursor, tc.want)
		}
		if len(got) != 1 || got[0] != tc.want {
			t.Fatalf("hover at %d,%d shapes = %v, want [%v]", tc.x, tc.y, got, tc.want)
		}
	}

	got = nil
	px, py := s.contentAt(100, 100, frameW, frameH)
	if err := s.hoverCursor(100, 100, px, py); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeDefault {
		t.Fatalf("off-thumb cursor = %v, want default", s.cursor)
	}
	if len(got) != 1 || got[0] != ebiten.CursorShapeDefault {
		t.Fatalf("off-thumb shapes = %v, want [default]", got)
	}

	s.dragAxis = dragHorizontal
	got = nil
	px, py = s.contentAt(198, 10, frameW, frameH)
	if err := s.hoverCursor(198, 10, px, py); err != nil {
		t.Fatal(err)
	}
	if s.cursor != ebiten.CursorShapeEWResize {
		t.Fatalf("cursor during horizontal drag = %v, want %v", s.cursor, ebiten.CursorShapeEWResize)
	}
}
