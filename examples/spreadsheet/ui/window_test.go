package ui

import (
	"context"
	"testing"
)

func TestWindowReplacesAndKeepsSelection(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	app.selectRect(2, 3, 5, 7)
	settle(t, app)
	if !app.win.Cover(Area{2, 3, 5, 7}) {
		t.Fatalf("selection outside window %+v", app.win)
	}

	app.page.SetScrollOffset(0, 5000)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.win.R0 <= 5 {
		t.Fatalf("window did not move: %+v", app.win)
	}

	if s := app.selection(); s != (Selection{2, 3, 5, 7}) {
		t.Fatalf("selection changed: %+v", s)
	}

	if _, ok := cellBox(app.View(), 3, 4); ok {
		t.Fatal("an offscreen cell is still rendered")
	}

	// Scrolling back shows the selected cells again with stable boxes.
	app.page.SetScrollOffset(0, 0)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if b, ok := cellBox(app.View(), 3, 4); !ok || b.X != cellX(4) || b.Y != cellY(3) {
		t.Fatalf("cell box = %+v ok=%v", b, ok)
	}

	if app.View().SelRect == nil {
		t.Fatal("selection rectangle missing")
	}
}
