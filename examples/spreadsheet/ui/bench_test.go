package ui

import (
	"context"
	"testing"
)

// benchApp builds a screen with the first window's tiles already loaded.
func benchApp(b *testing.B) *App {
	b.Helper()

	fake := newFake()
	app, err := New(Options{Backend: fake, Width: 1200, Height: 800})
	if err != nil {
		b.Fatal(err)
	}

	b.Cleanup(func() { _ = app.Close() })
	app.win = computeWindow("s1", 0, 0, 1200, 800, 200, 20, FetchOver)
	cells, err := fake.Range("s1", app.win.Area())
	if err != nil {
		b.Fatal(err)
	}

	app.tiles.put("s1", app.win.Area(), cells, 0)
	app.refresh()
	if err := app.page.Redraw(context.Background()); err != nil {
		b.Fatal(err)
	}

	return app
}

// BenchmarkRedrawScrollStep is the cost of one wheel step: the chrome
// moves, the rendered window is unchanged.
func BenchmarkRedrawScrollStep(b *testing.B) {
	app := benchApp(b)
	ctx := context.Background()
	base := app.scrollY

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.page.SetScrollOffset(0, base+(i%100)*48)
		app.page.Testing().StepScrollWindow()
		if err := app.page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRedrawDense fills every rendered cell with text.
func BenchmarkRedrawDense(b *testing.B) {
	app := benchApp(b)
	ctx := context.Background()
	a := app.win.Area()
	cells := make([]Cell, 0, a.Count())
	for i := 0; i < a.Count(); i++ {
		cells = append(cells, Cell{Raw: "value 123", Display: "value 123"})
	}

	app.tiles.put("s1", a, cells, 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app.page.SetScrollOffset(0, (i%100)*48)
		app.page.Testing().StepScrollWindow()
		if err := app.page.Redraw(ctx); err != nil {
			b.Fatal(err)
		}
	}
}
