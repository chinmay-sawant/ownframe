package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestPressEdge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                string
		down, released, was bool
		want                bool
	}{
		{"fresh press", true, false, false, true},
		{"held press", true, false, true, false},
		{"fast pair between frames", false, true, false, true},
		{"release after seen press", false, true, true, false},
		{"idle frame", false, false, false, false},
	}

	for _, c := range cases {
		if got := pressEdge(c.down, c.released, c.was); got != c.want {
			t.Fatalf("%s: pressEdge(%v, %v, %v) = %v", c.name,
				c.down, c.released, c.was, got)
		}
	}
}

func TestFastClickBetweenFramesStillClicks(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	clicked, err := s.pointerMouse(100, 100, 100, 100, false, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !clicked || app.presses != 1 || app.clicks != 1 || app.releases != 1 {
		t.Fatalf("events = %v/%d/%d/%d, want one press, click, release", clicked,
			app.presses, app.clicks, app.releases)
	}
}

func TestReleaseAfterSeenPressDoesNotClickAgain(t *testing.T) {
	app := &pointerRouteScreen{fakeScreen: fakeScreen{
		width: 200, height: 200,
		boxes: []layout.Box{{W: 500, H: 500}},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}

	if _, err := s.pointerMouse(100, 100, 100, 100, true, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := s.pointerMouse(100, 100, 100, 100, false, true, false); err != nil {
		t.Fatal(err)
	}
	if app.clicks != 1 {
		t.Fatalf("clicks = %d, want 1", app.clicks)
	}
}
