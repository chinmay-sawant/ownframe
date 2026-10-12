package snip_test

import (
	"testing"
)

func TestSelectClampsRect(t *testing.T) {
	t.Parallel()

	app, ctx := newTestApp(t, fakeShot(160, 100))
	app.SetForm("x", "-5")
	app.SetForm("y", "3")
	app.SetForm("w", "99999")
	app.SetForm("h", "7")

	if err := app.Handle(ctx, "select"); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.SelX != 0 || view.SelY != 3 || view.SelW != 160 || view.SelH != 7 {
		t.Fatalf("rect = %d,%d %dx%d, want 0,3 160x7",
			view.SelX, view.SelY, view.SelW, view.SelH)
	}
}

func TestClearResets(t *testing.T) {
	t.Parallel()

	app, ctx := newTestApp(t, fakeShot(160, 100))
	app.SetForm("w", "10")
	app.SetForm("h", "10")

	if err := app.Handle(ctx, "select"); err != nil {
		t.Fatal(err)
	}

	if err := app.Handle(ctx, "clear"); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if view.SelX != 0 || view.SelY != 0 || view.SelW != 160 || view.SelH != 100 {
		t.Fatalf("rect = %+v, want full capture", view)
	}

	if view.Status != "cleared" {
		t.Fatalf("status = %q, want cleared", view.Status)
	}
}
