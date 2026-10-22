package ui

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestTickSurvivesBitmapFallback(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)

	page, err := ownframe.New(ownframe.Config{
		HTML:   `<html><body><div style="mix-blend-mode:multiply">blend</div></body></html>`,
		Width:  400,
		Height: 300,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := page.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if page.Testing().Display() != nil {
		t.Skip("page is replayable; cannot exercise the bitmap fallback")
	}

	app.page = page
	app.state.h = handles{}
	app.mail.putSummary(app.currentGen(), sample("cpu", 0.5))

	if err := app.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
}
