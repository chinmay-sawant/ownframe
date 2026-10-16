package watch

import (
	"context"
	"testing"
	"time"
)

func TestDisplayReplayable(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw() failed: %v", err)
	}
	if app.page.Display() == nil {
		t.Fatal("expected watch display to be replayable without bitmap fallback")
	}

	if err := app.Tick(context.Background()); err != nil {
		t.Fatalf("Tick() failed: %v", err)
	}

	p := &app.parts
	if len(p.handHour.ops) == 0 {
		t.Fatal("expected handHour ops to be bound")
	}
	if len(p.handMinute.ops) == 0 {
		t.Fatal("expected handMinute ops to be bound")
	}
	if len(p.handSecond.ops) == 0 {
		t.Fatal("expected handSecond ops to be bound")
	}
	if len(p.tourbillon.ops) == 0 {
		t.Fatal("expected tourbillon ops to be bound")
	}
	if len(p.chronoMin.ops) == 0 {
		t.Fatal("expected chronoMin ops to be bound")
	}
	if len(p.powerHand.ops) == 0 {
		t.Fatal("expected powerHand ops to be bound")
	}

	t0 := app.now()
	app.now = func() time.Time { return t0 }
	_ = app.Tick(context.Background())
	b0 := p.tourbillon.ops[0].Transform().B

	app.now = func() time.Time { return t0.Add(5 * time.Second) }
	_ = app.Tick(context.Background())
	b1 := p.tourbillon.ops[0].Transform().B

	if b0 == b1 {
		t.Fatalf("expected tourbillon rotation to change over time, but remained %f", b0)
	}
}
