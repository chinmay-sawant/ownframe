package watch

import (
	"context"
	"testing"
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
}
