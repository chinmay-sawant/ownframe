package inputlab

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func newTest(t *testing.T) (*App, context.Context) {
	t.Helper()
	app, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	return app, ctx
}

func clickID(t *testing.T, app *App, ctx context.Context, id string) {
	t.Helper()
	for _, b := range app.Boxes() {
		if b.ID == id && b.W > 0 && b.H > 0 {
			if err := app.Click(ctx, b.X+b.W/2, b.Y+b.H/2); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("no box %s", id)
}

func boxByID(t *testing.T, app *App, id string) ownframe.Box {
	t.Helper()
	for _, b := range app.Boxes() {
		if b.ID == id {
			return b
		}
	}
	t.Fatalf("no box %s", id)
	return ownframe.Box{}
}

func TestBindWritesStruct(t *testing.T) {
	app, ctx := newTest(t)
	clickID(t, app, ctx, "b-name")
	if err := app.Type(ctx, "Ada"); err != nil {
		t.Fatal(err)
	}
	if app.View().BName != "Ada" {
		t.Fatalf("BName=%q", app.View().BName)
	}
}

func TestLockedShowsMessage(t *testing.T) {
	app, ctx := newTest(t)
	clickID(t, app, ctx, "l-name")
	if err := app.Type(ctx, "x"); err != nil {
		t.Fatalf("locked edit errored: %v", err)
	}
	if got := app.FormValue("l-name"); got != "locked" {
		t.Fatalf("l-name=%q", got)
	}
	if got := app.View().LStatus; got != lockedMsg {
		t.Fatalf("LStatus=%q", got)
	}
}

func TestSendSnapshot(t *testing.T) {
	app, ctx := newTest(t)
	clickID(t, app, ctx, "f-email")
	if err := app.Type(ctx, "a@ex.com"); err != nil {
		t.Fatal(err)
	}
	clickID(t, app, ctx, "f-send")
	if got := app.View().Status; got == "" {
		t.Fatal("empty status")
	}
}

func TestMultilineTextareaSection(t *testing.T) {
	app, ctx := newTest(t)
	if err := app.Page().Form().Focus(ctx, "ml-editor"); err != nil {
		t.Fatal(err)
	}
	before := app.FormValue("ml-editor")
	if !strings.Contains(before, "\n\n") {
		t.Fatalf("ml-editor missing prefilled paragraphs: %q", before)
	}
	if err := app.Page().Edit().KeyDown(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Edit().KeyUp(ctx, "enter"); err != nil {
		t.Fatal(err)
	}
	if got := app.FormValue("ml-editor"); got != before+"\n" {
		t.Fatalf("after enter got %q, want trailing newline", got)
	}
	if !strings.Contains(app.View().MLReadout, "focus=ml-editor") ||
		!strings.Contains(app.View().MLReadout, "newlines=4") {
		t.Fatalf("MLReadout = %q", app.View().MLReadout)
	}
}
