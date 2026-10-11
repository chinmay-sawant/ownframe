package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestContentClickAndScrollbarDragKeepTheirEvents(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	move := func(x, y int, down, released bool) {
		t.Helper()
		if _, err := s.pointerMouse(x, y, float64(x), float64(y), down, released, false); err != nil {
			t.Fatal(err)
		}
	}

	move(100, 100, true, false)
	move(100, 100, false, true)
	if app.presses != 1 || app.clicks != 1 || app.releases != 1 {
		t.Fatalf("content click events = %d/%d/%d, want 1/1/1", app.presses, app.clicks, app.releases)
	}

	move(198, 10, true, false)
	move(198, 100, true, false)
	if s.scrollY == 0 || app.presses != 1 || app.clicks != 1 || app.drags != 0 {
		t.Fatalf("thumb drag offset/events = %d/%d/%d/%d", s.scrollY, app.presses, app.clicks, app.drags)
	}
	move(198, 100, false, true)
	if app.clicks != 1 || app.releases != 2 {
		t.Fatalf("thumb release events = clicks %d, releases %d", app.clicks, app.releases)
	}
	move(10, 198, true, false)
	move(100, 198, true, false)
	if s.scrollX == 0 || app.presses != 1 || app.clicks != 1 || app.drags != 0 {
		t.Fatalf("horizontal thumb offset/events = %d/%d/%d/%d", s.scrollX, app.presses, app.clicks, app.drags)
	}
	move(100, 198, false, true)

	_, wheelY := panScroll(0, s.scrollY, 0, -1, 500, 500, 200, 200)
	if wheelY <= s.scrollY {
		t.Fatalf("page wheel offset = %d after thumb offset %d", wheelY, s.scrollY)
	}
}
