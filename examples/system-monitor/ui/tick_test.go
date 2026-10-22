package ui

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// newTestApp builds an app without polls and redraws once.
func newTestApp(t *testing.T, src Source, store Store) *App {
	t.Helper()

	app, err := New(context.Background(), Config{
		Source: src,
		Store:  store,
		noPump: true,
		Width:  1024,
		Height: 700,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = app.Close() })

	if err := app.page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	return app
}

func TestTickPaintsBarsAndRebinds(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)
	ctx := context.Background()

	if app.page.Display() == nil {
		t.Fatal("page fell back to the bitmap path; tick cannot paint")
	}

	t.Logf("display ops=%d boxes=%d", len(app.page.Display().Ops), len(app.page.Boxes()))

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	bars := app.state.h.bars["cpu"]
	if len(bars) == 0 {
		t.Fatal("no cpu bars bound")
	}

	flat := maxBar(bars)

	app.mail.putSummary(app.currentGen(), sample("cpu", 0.9))

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if got := maxBar(bars); got <= flat {
		t.Fatalf("tallest bar %v did not grow from %v", got, flat)
	}

	if err := app.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.state.h.gen != app.page.Generation() {
		t.Fatal("handles not reacquired after redraw")
	}

	fresh := app.state.h.bars["cpu"]
	if len(fresh) == 0 || maxBar(fresh) <= flat {
		t.Fatal("painted value lost after redraw")
	}
}

// maxBar returns the tallest bar's height.
func maxBar(bars []*ownframe.DisplayOp) float64 {
	h := 0.0

	for _, b := range bars {
		h = max(h, b.H)
	}

	return h
}
