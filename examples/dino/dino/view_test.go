package dino

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestTouchViewFitsThePhone(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()
	app.page.SetSize(881, 396)

	if err := app.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	tick(t, app, clock, time.Second/60)

	if app.parts.ground == nil {
		t.Fatal("the ground was not bound")
	}

	want := fitView(881, 396)
	pp := app.page.Display().PointsPerPixel

	if got := app.parts.ground.Y / pp; math.Abs(got-((groundY-2)*want.scale+want.oy)) > 1 {
		t.Fatalf("ground y = %v", got)
	}

	if got := app.parts.ground.W / pp; math.Abs(got-sceneW*want.scale) > 1 {
		t.Fatalf("ground w = %v", got)
	}
}

func TestDesktopViewIsIdentity(t *testing.T) {
	app, clock := newTestApp(t)

	tick(t, app, clock, time.Second/60)

	if app.parts.ground == nil {
		t.Fatal("the ground was not bound")
	}

	pp := app.page.Display().PointsPerPixel

	if got := app.parts.ground.Y / pp; math.Abs(got-(groundY-2)) > 1 {
		t.Fatalf("ground y = %v, want %v", got, float64(groundY-2))
	}
}
