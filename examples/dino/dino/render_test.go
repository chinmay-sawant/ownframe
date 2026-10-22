package dino

import (
	"math"
	"strings"
	"testing"
	"time"
)

func TestPageReplaysWithoutAButton(t *testing.T) {
	app, _ := newTestApp(t)

	if app.page.Display() == nil {
		t.Fatal("the page fell back to the bitmap path")
	}

	html := buildHTML()

	for _, bad := range []string{"<button", "data-action", "<input"} {
		if strings.Contains(html, bad) {
			t.Fatalf("the game page has %q", bad)
		}
	}
}

func TestTickPaintsTheObstacle(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")

	app.game.obstacles = []obstacle{{kind: cactusSmall, x: 500, w: 22, h: 32}}
	tick(t, app, clock, time.Second/60)

	op := app.parts.slots[0][partSmall]
	if op == nil {
		t.Fatal("the small cactus was not bound")
	}

	if op.W <= 0 || op.H <= 0 {
		t.Fatal("the small cactus is hidden")
	}

	pp := app.page.Display().PointsPerPixel
	if got := op.X / pp; math.Abs(got-500) > 8 {
		t.Fatalf("cactus x = %v, want about 500", got)
	}
}

func TestTickHidesTheEmptySlots(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	tick(t, app, clock, time.Second/60)

	for slot := 1; slot < slotMax; slot++ {
		for i, op := range app.parts.slots[slot] {
			if op != nil && op.Alpha != 0 {
				t.Fatalf("slot %d part %d is visible with no obstacle", slot, i)
			}
		}
	}
}

func TestTickMovesTheDinoPose(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	tick(t, app, clock, time.Second/60)

	pp := app.page.Display().PointsPerPixel
	top := func() float64 { return app.parts.dino[1].Y / pp }

	ground := top()

	press(t, app, "w")
	tick(t, app, clock, time.Second/60)

	if got := top(); got >= ground {
		t.Fatalf("the dino pose did not rise: %v then %v", ground, got)
	}
}
