package board

import (
	"context"
	"testing"
)

// TestStatusGrowthShowsTail binds at feed 0, then rotates until row 1 walks
// from On time to Delayed 10. The padded tail tiles must take the new glyphs
// with the generation untouched and the cascade settled.
func TestStatusGrowthShowsTail(t *testing.T) {
	t.Parallel()

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

	gen := app.Page().Generation()

	// Row 1 starts on On time and steps once every fifth feed; feed 19 is
	// the first feed where it reads Delayed 10.
	for app.feed < 19 {
		forceRotation(app)

		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}

		for i := 0; i < 200 && !flapsHold(app); i++ {
			if err := app.Tick(ctx); err != nil {
				t.Fatal(err)
			}
		}
	}

	if got := app.Page().Generation(); got != gen {
		t.Fatalf("rotations redrew the page: generation %d -> %d", gen, got)
	}

	if got := app.view.Rows[1].Status; got != "Delayed 10" {
		t.Fatalf("row 1 status = %q, want Delayed 10", got)
	}

	fl := app.flip.cells["t-1-stat-9"]
	if fl == nil {
		t.Fatal("no flap bound for the tail tile t-1-stat-9")
	}

	if fl.want != "0" {
		t.Fatalf("tail want = %q, want 0", fl.want)
	}

	if fl.op.Text != "0" {
		t.Fatalf("tail shows %q, want 0", fl.op.Text)
	}
}
