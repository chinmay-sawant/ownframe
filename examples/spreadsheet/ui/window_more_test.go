package ui

import (
	"context"
	"testing"
)

func TestCtrlArrowDownWithTrackedModifier(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	ctx := context.Background()
	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyDown(ctx, "arrowdown"); err != nil {
		t.Fatal(err)
	}

	if s := app.selection(); s.ActiveR != 199 {
		t.Fatalf("active row = %d", s.ActiveR)
	}
}

func TestEditorTextSurvivesViewportReplacement(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	ctx := context.Background()
	if err := app.onType(ctx, "abc"); err != nil {
		t.Fatal(err)
	}

	// Scroll far down: A1 leaves the rendered window.
	app.page.Scroll().SetScrollOffset(0, 5000)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.edit == nil || app.edit.Text != "abc" {
		t.Fatalf("editor lost: %+v", app.edit)
	}

	if app.View().Editor != nil {
		t.Fatal("offscreen editor still rendered")
	}

	// Scroll back: the buffer and the caret return with the cell.
	app.page.Scroll().SetScrollOffset(0, 0)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if v := app.View().Editor; v == nil || v.Before != "abc" {
		t.Fatalf("editor view = %+v", v)
	}
}

func TestHorizontalWindow(t *testing.T) {
	app := newTestApp(t, newFake())
	settle(t, app)

	app.page.Scroll().SetScrollOffset(1000, 0)
	app.page.Testing().StepScrollWindow()
	flush(t, app)
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.win.C0 == 0 {
		t.Fatalf("window did not move horizontally: %+v", app.win)
	}

	if b, ok := cellBox(app.View(), 0, app.win.C0); !ok || b.X != cellX(app.win.C0) {
		t.Fatalf("stable column coordinates: %+v ok=%v", b, ok)
	}
}
