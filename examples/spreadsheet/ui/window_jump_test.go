package ui

import (
	"context"
	"testing"
)

func TestKeyboardJumpToOffscreenCell(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	ctx := context.Background()
	if err := app.onKeyDown(ctx, "ctrl+end"); err != nil {
		t.Fatal(err)
	}

	flush(t, app)

	s := app.selection()
	if s.ActiveR != 150 || s.ActiveC != 15 {
		t.Fatalf("active = %d,%d", s.ActiveR, s.ActiveC)
	}

	req, ok := app.page.Testing().TakeScroll()
	if !ok || !req.Absolute {
		t.Fatal("no scroll request for the offscreen cell")
	}

	app.page.Scroll().SetScrollOffset(req.X, req.Y)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !app.win.Cover(Area{s.ActiveR, s.ActiveC, s.ActiveR, s.ActiveC}) {
		t.Fatalf("window %+v misses the active cell", app.win)
	}

	if _, ok := cellBox(app.View(), 150, 15); !ok {
		t.Fatal("far cell not rendered after the jump")
	}
}
