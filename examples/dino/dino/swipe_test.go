package dino

import (
	"context"
	"testing"
	"time"
)

func TestSwipeUpJumps(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()

	swipe(t, app, 0, -60)

	if app.game.phase != running {
		t.Fatal("a swipe up did not start the run")
	}

	tick(t, app, clock, time.Second/60)
	swipe(t, app, 0, -60)

	if app.game.onGround {
		t.Fatal("a swipe up did not jump")
	}
}

func TestSwipeDownDucks(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()

	swipe(t, app, 0, -60)
	tick(t, app, clock, time.Second/60)
	swipe(t, app, 0, 60)

	if !app.game.ducking {
		t.Fatal("a swipe down did not duck")
	}

	tick(t, app, clock, duckHold+time.Second/60)

	if app.game.ducking {
		t.Fatal("the duck outlasted its hold")
	}
}

func TestSwipeUpCancelsDuck(t *testing.T) {
	app, clock := newTestApp(t)
	app.BindTouch()

	swipe(t, app, 0, -60)
	tick(t, app, clock, time.Second/60)
	swipe(t, app, 0, 60)
	swipe(t, app, 0, -60)

	if app.game.ducking || app.game.onGround {
		t.Fatal("a swipe up did not cancel the duck and jump")
	}
}

func TestSwipeSidewaysIgnored(t *testing.T) {
	app, _ := newTestApp(t)
	app.BindTouch()

	swipe(t, app, 60, 10)

	if app.game.phase != ready || app.game.ducking || !app.game.onGround {
		t.Fatal("a sideways swipe changed the game")
	}
}

// swipe sends a one-finger swipe through the page handlers.
func swipe(t *testing.T, app *App, dx, dy float64) {
	t.Helper()

	if err := app.page.Testing().Swipe(context.Background(), dx, dy); err != nil {
		t.Fatal(err)
	}
}
