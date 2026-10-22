package window

import (
	"context"
	"testing"
)

func TestReloadResolvesHoverFromTheCursor(t *testing.T) {
	t.Parallel()

	app := newReloadScreen()
	app.left = 1

	s := newShell(context.Background(), app)
	s.cursorX, s.cursorY = 10, 20

	if err := s.pollReload(); err != nil {
		t.Fatal(err)
	}

	if app.hovers != 1 {
		t.Fatalf("hovers = %d, want 1", app.hovers)
	}

	if app.hoverX != 10 || app.hoverY != 20 {
		t.Fatalf("hover = %v, %v", app.hoverX, app.hoverY)
	}
}
