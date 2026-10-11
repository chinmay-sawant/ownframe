package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestHeldPressDoesNotStartThumbDrag(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, mouseDown: true}
	if _, err := s.pointerMouse(198, 10, 198, 10, true, false, false); err != nil {
		t.Fatal(err)
	}
	if s.dragAxis != dragNone || app.clicks != 0 {
		t.Fatalf("held press grabbed thumb or clicked page: axis %d, clicks %d", s.dragAxis, app.clicks)
	}
}
