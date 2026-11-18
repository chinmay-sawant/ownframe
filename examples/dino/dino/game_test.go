package dino

import (
	"testing"
	"time"
)

func TestRunMovesObstaclesAndScores(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")

	app.game.obstacles = []obstacle{{kind: cactusSmall, x: 800, w: 22, h: 32}}

	for range 60 {
		tick(t, app, clock, time.Second/60)
	}

	if app.game.score == 0 {
		t.Fatal("no score after a second")
	}

	if app.game.speed <= 330 {
		t.Fatalf("speed = %v, want more than 330", app.game.speed)
	}

	if len(app.game.obstacles) == 0 || app.game.obstacles[0].x >= 800 {
		t.Fatal("the obstacle did not move left")
	}
}

func TestBestScoreTracksActiveRun(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	app.game.obstacles = []obstacle{{kind: cactusSmall, x: 800, w: 22, h: 32}}

	for range 60 {
		tick(t, app, clock, time.Second/60)
	}

	if app.game.phase != running {
		t.Fatal("run ended before the score could advance")
	}
	if app.game.high != app.game.score {
		t.Fatalf("best score = %d, current score = %d", app.game.high, app.game.score)
	}
}

func TestJumpRisesAndLands(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")
	press(t, app, "space")
	release(t, app, "space")

	for range 12 {
		tick(t, app, clock, time.Second/60)
	}

	if app.game.feet <= 0 {
		t.Fatal("the jump did not leave the ground")
	}

	for range 120 {
		tick(t, app, clock, time.Second/60)
	}

	if !app.game.onGround || app.game.feet != 0 {
		t.Fatalf("dino did not land: feet = %v", app.game.feet)
	}
}

func TestCrashThenRestart(t *testing.T) {
	app, clock := newTestApp(t)
	press(t, app, "space")

	app.game.distance = 140 * 18
	app.game.obstacles = []obstacle{{kind: cactusBig, x: dinoX - 2, w: 30, h: 48}}
	tick(t, app, clock, time.Second/60)

	if app.game.phase != over {
		t.Fatal("the crash did not end the run")
	}

	if app.game.high != 140 {
		t.Fatalf("best = %d, want 140", app.game.high)
	}

	press(t, app, "space")

	if app.game.phase != running {
		t.Fatal("space did not start a new run")
	}

	if app.game.score != 0 || len(app.game.obstacles) != 0 {
		t.Fatal("the new run did not reset")
	}

	if app.game.high != 140 {
		t.Fatal("the new run lost the best score")
	}
}
