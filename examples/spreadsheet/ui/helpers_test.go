package ui

import (
	"context"
	"testing"
	"time"
)

// newTestApp builds a screen over a fake backend at a fixed size.
func newTestApp(t *testing.T, b *fakeBackend) *App {
	t.Helper()

	app, err := New(Options{Backend: b, Width: 1000, Height: 700, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = app.Close() })

	return app
}

// flush drains worker results until the worker is idle and the result
// channel is empty.
func flush(t *testing.T, app *App) {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if err := app.tick(context.Background()); err != nil {
			t.Fatal(err)
		}

		if app.work.idle() && len(app.work.out) == 0 {
			return
		}

		time.Sleep(time.Millisecond)
	}

	t.Fatal("worker did not settle")
}

// settle flushes the worker, applies a queued page scroll the way the
// window does, and redraws.
func settle(t *testing.T, app *App) {
	t.Helper()

	flush(t, app)
	if req, ok := app.page.Testing().TakeScroll(); ok {
		app.page.Scroll().SetScrollOffset(req.X, req.Y)
		app.page.Testing().StepScrollWindow()
		flush(t, app)
	}

	if err := app.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
}

// cellBox returns the view box of one cell.
func cellBox(v View, r, c int) (Box, bool) {
	want := "cell:" + itoa(r) + ":" + itoa(c)
	for _, b := range v.Cells {
		if string(b.Action) == want {
			return b, true
		}
	}

	return Box{}, false
}

// clickAction clicks one toolbar button by its action name.
func clickAction(t *testing.T, app *App, action string) {
	t.Helper()

	for _, b := range app.View().Toolbar {
		if string(b.Action) == action {
			clickBox(t, app, b)

			return
		}
	}

	t.Fatalf("toolbar action %q missing", action)
}

// clickBox clicks the centre of one view box.
func clickBox(t *testing.T, app *App, b Box) {
	t.Helper()

	if err := app.page.Click(context.Background(), float64(b.X+b.W/2), float64(b.Y+b.H/2)); err != nil {
		t.Fatal(err)
	}
}
