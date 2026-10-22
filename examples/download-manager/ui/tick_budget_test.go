package ui

import (
	"context"
	"testing"
)

func TestTickDrainBudget(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 1, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.view.Stats.Applied = 0

	for i := 0; i < drainBudget+10; i++ {
		row := running("j1", int64(i), 100)
		back.send(Update{Kind: UpdateProgress, Row: &row})
	}

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.view.Stats.Applied != drainBudget {
		t.Fatalf("applied %d updates, want %d", app.view.Stats.Applied, drainBudget)
	}

	if back.queued() != 10 {
		t.Fatalf("queue left with %d, want 10", back.queued())
	}
}
