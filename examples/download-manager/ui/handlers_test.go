package ui

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSubmitAddsJob(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	box, ok := findBox(app.Boxes(), "url")
	if !ok {
		t.Fatal("no url field")
	}

	if err := app.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatalf("focus click: %v", err)
	}

	app.page.Form().SetFormValue("url", "https://example.com/enter.bin")

	if err := app.onSubmit(ctx); err != nil {
		t.Fatalf("onSubmit: %v", err)
	}

	if len(back.added) != 1 || back.added[0].URL != "https://example.com/enter.bin" {
		t.Fatalf("add = %+v", back.added)
	}
}

func TestChangeSetsDestination(t *testing.T) {
	app, _ := newTestApp(t)
	ctx := context.Background()
	picked := filepath.Join(t.TempDir(), "pick.bin")

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	app.page.Form().SetFormValue("dest-file", picked)

	if err := app.onChange(ctx, ownframe.Box{ID: "dest-file"}); err != nil {
		t.Fatalf("onChange: %v", err)
	}

	if got := app.page.Form().FormValue("dest"); got != filepath.Dir(picked) {
		t.Fatalf("dest = %q, want %q", got, filepath.Dir(picked))
	}

	if got := app.page.Form().FormValue("dest-file"); got != "" {
		t.Fatalf("dest-file = %q, want empty", got)
	}
}
