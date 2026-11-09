package board

import (
	"context"
	"testing"
	"time"
)

// TestWantFrameAsksForTheClock pins the idle cadence: with the flaps settled
// the board asks for the next clock second, not a frame every 100 ms.
func TestWantFrameAsksForTheClock(t *testing.T) {
	ctx := context.Background()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	at := app.Page().Testing().NextFrameAt()
	if at.IsZero() {
		t.Fatal("the tick did not declare a frame")
	}

	// Allow a small scheduling margin: the tick sets the deadline one second
	// out, and the assertion itself takes a little time on a loaded runner.
	if d := time.Until(at); d <= 0 || d > time.Second+100*time.Millisecond {
		t.Fatalf("idle deadline %v, want within a second", d)
	}
}

// TestWantFrameSpinsEveryFrame pins the animation cadence: while a flap is
// moving the board asks for the next frame at once.
func TestWantFrameSpinsEveryFrame(t *testing.T) {
	ctx := context.Background()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		runTick(t, ctx, app)
	}

	forceRotation(app)

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.Page().Testing().NextFrameAt()

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	at := app.Page().Testing().NextFrameAt()
	if at.IsZero() {
		t.Fatal("the spinning tick did not declare a frame")
	}

	if !at.Before(time.Now()) {
		t.Fatalf("spinning deadline %v, want now or earlier", at)
	}
}
