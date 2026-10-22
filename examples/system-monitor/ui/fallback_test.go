package ui

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTickSurvivesBitmapFallback(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)

	replayable, err := ownframe.New(ownframe.Config{
		HTML:   `<html><body><div style="mix-blend-mode:multiply">blend</div></body></html>`,
		Width:  400,
		Height: 300,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := replayable.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if replayable.Display() == nil {
		t.Fatal("separable blend fell back to the bitmap")
	}

	if replayable.Image() != nil {
		t.Fatal("separable blend required the bitmap fallback")
	}

	app.page = replayable
	app.state.h = handles{}
	app.mail.putSummary(app.currentGen(), sample("cpu", 0.5))

	if err := app.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

	fallback, err := ownframe.New(ownframe.Config{
		HTML: `<html><body><div style="width:60px;height:40px;background:#123456;` +
			`border-radius:50% / 20%"></div></body></html>`,
		Width:  400,
		Height: 300,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := fallback.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if fallback.Display() != nil {
		t.Fatal("elliptical border was replayed")
	}

	if fallback.Image() == nil {
		t.Fatal("no fallback bitmap")
	}

	if len(fallback.PNG()) == 0 {
		t.Fatal("fallback page has no PNG")
	}

	app.page = fallback
	app.state.h = handles{}
	app.mail.putSummary(app.currentGen(), sample("cpu", 0.5))

	if err := app.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}
