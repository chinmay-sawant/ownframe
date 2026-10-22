package flappy

import (
	"testing"
	"time"
)

func TestPaintPageStaysReplayable(t *testing.T) {
	app, _ := newTestApp(t)

	if app.page.Display() == nil {
		t.Fatal("the page fell back to the bitmap path")
	}

	if app.page.Image() != nil {
		t.Fatal("the page has a bitmap as well as a display list")
	}
}

func TestPaintTickMovesThePipe(t *testing.T) {
	app, clock := newTestApp(t)
	tick(t, app, clock, time.Second/60)
	press(t, app, "space")

	app.game.pipes = []pipe{{x: 400, gapY: sceneH / 2, gap: gapStart}}
	tick(t, app, clock, time.Second/60)

	op := app.parts.pipes[0][0]
	if op == nil {
		t.Fatal("the pipe top was not bound")
	}

	before := op.X
	tick(t, app, clock, time.Second/60)

	if got := op.X; got >= before {
		t.Fatalf("the pipe did not move left: %v then %v", before, got)
	}
}

func TestPaintTickHidesEmptyPipeSlots(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	app.game.pipes = []pipe{{x: 400, gapY: sceneH / 2, gap: gapStart}}
	tick(t, app, clock, time.Second/60)

	for slot := 1; slot < pipeSlots; slot++ {
		for i, op := range app.parts.pipes[slot] {
			if op != nil && (op.W != 0 || op.H != 0) {
				t.Fatalf("slot %d part %d is visible with no pipe", slot, i)
			}
		}
	}
}

func TestPaintFlapRaisesTheBird(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	tick(t, app, clock, time.Second/60)

	op := app.parts.bird[0]
	if op == nil {
		t.Fatal("the bird body was not bound")
	}

	pp := app.page.Display().PointsPerPixel
	top := func() float64 { return op.Y / pp }

	before := top()
	press(t, app, "space")
	tick(t, app, clock, time.Second/60)

	if got := top(); got >= before {
		t.Fatalf("the bird did not rise: %v then %v", before, got)
	}
}
