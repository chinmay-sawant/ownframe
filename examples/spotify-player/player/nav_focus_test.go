package player

import (
	"context"
	"testing"
)

func TestNavigationFocusRetention(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	ctx := context.Background()

	app.Page().SetScrollOffset(0, 42)
	if err := app.Page().Focus(ctx, "q"); err != nil {
		t.Fatalf("Focus: %v", err)
	}
	redraw(t, app)

	click(t, app, "nav-search")
	redraw(t, app)

	app.Page().SetScrollOffset(0, 7)
	if err := app.Page().Focus(ctx, "q"); err != nil {
		t.Fatalf("Focus: %v", err)
	}
	redraw(t, app)

	click(t, app, "nav-home")
	redraw(t, app)

	if x, y := app.Page().ScrollOffset(); x != 0 || y != 42 {
		t.Fatalf("home scroll = %d,%d, want 0,42", x, y)
	}
	if got := app.Page().FocusID(); got != "q" {
		t.Fatalf("home focus = %q, want q", got)
	}

	click(t, app, "nav-search")
	redraw(t, app)

	if x, y := app.Page().ScrollOffset(); x != 0 || y != 7 {
		t.Fatalf("search scroll = %d,%d, want 0,7", x, y)
	}
	if got := app.Page().FocusID(); got != "q" {
		t.Fatalf("search focus = %q, want q", got)
	}

	redraw(t, app)
}

func TestBackAndForwardRetainFocusScroll(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)

	ctx := context.Background()

	app.Page().SetScrollOffset(0, 42)
	_ = app.Page().Focus(ctx, "q")
	redraw(t, app)

	click(t, app, "nav-search")
	redraw(t, app)

	app.Page().SetScrollOffset(0, 7)
	_ = app.Page().Focus(ctx, "q")
	redraw(t, app)

	click(t, app, "back")
	redraw(t, app)

	if x, y := app.Page().ScrollOffset(); x != 0 || y != 42 {
		t.Fatalf("back scroll = %d,%d, want 0,42", x, y)
	}
	if got := app.Page().FocusID(); got != "q" {
		t.Fatalf("back focus = %q, want q", got)
	}

	click(t, app, "forward")
	redraw(t, app)

	if x, y := app.Page().ScrollOffset(); x != 0 || y != 7 {
		t.Fatalf("forward scroll = %d,%d, want 0,7", x, y)
	}

	redraw(t, app)
}
