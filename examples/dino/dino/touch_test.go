package dino

import (
	"context"
	"testing"
	"time"
)

func TestTapStartsAndJumps(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()

	tapScene(t, app)

	if app.game.phase != running {
		t.Fatal("a tap did not start the run")
	}

	tick(t, app, clock, time.Second/60)
	tapScene(t, app)

	if app.game.onGround {
		t.Fatal("a tap did not jump")
	}

	for range 30 {
		tick(t, app, clock, time.Second/60)
	}

	if app.game.jumpHeld || app.tapped {
		t.Fatal("the tap jump stayed held")
	}
}

func TestTapRestartsAfterACrash(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()

	tapScene(t, app)
	app.game.phase = over
	app.game.score = 90
	tapScene(t, app)
	tick(t, app, clock, time.Second/60)

	if app.game.phase != running || app.game.score != 0 || app.game.high != 90 {
		t.Fatal("a tap did not start a new run")
	}
}

func TestTouchHints(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()
	tick(t, app, clock, time.Second/60)

	if app.parts.start.Text != tapStartText || app.parts.keys.Text != tapKeysText {
		t.Fatal("the touch hints are wrong")
	}
}

// tapScene clicks the middle of the scene through the page handlers.
func tapScene(t *testing.T, app *App) {
	t.Helper()

	if err := app.page.Click(context.Background(), 450, 150); err != nil {
		t.Fatal(err)
	}
}
