package login_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestUndoRedoRestoresFieldValues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	if err := app.Type(ctx, "xy"); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "abxy" {
		t.Fatalf("typed = %q", got)
	}

	if err := app.Page().Edit().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "ab" {
		t.Fatalf("undo = %q", got)
	}

	if err := app.Page().Edit().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "" {
		t.Fatalf("second undo = %q", got)
	}

	if err := app.Page().Edit().Redo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "ab" {
		t.Fatalf("redo = %q", got)
	}

	if err := app.Page().Edit().Redo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "abxy" {
		t.Fatalf("second redo = %q", got)
	}
}

func TestUndoGraysTheButtonAgain(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Edit().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if app.View().Ready {
		t.Fatal("ready after undoing the password")
	}

	if got := buttonFill(t, app); got != disabledButton {
		t.Fatalf("undo fill = %+v, want %+v", got, disabledButton)
	}
}

func TestCopyPasteCutAndUndo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "ab"); err != nil {
		t.Fatal(err)
	}

	text, ok, err := app.Page().Edit().Copy(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ok || text != "ab" {
		t.Fatalf("copy = %q ok=%v", text, ok)
	}

	if err := app.Page().Edit().Paste(ctx, "xy"); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "abxy" {
		t.Fatalf("pasted = %q", got)
	}

	text, ok, err = app.Page().Edit().Cut(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ok || text != "abxy" {
		t.Fatalf("cut = %q ok=%v", text, ok)
	}

	if got := app.Page().Form().FormValue("email"); got != "" {
		t.Fatalf("after cut = %q", got)
	}

	if err := app.Page().Edit().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "abxy" {
		t.Fatalf("undo = %q", got)
	}
}

func TestSelectAllReplacesTheField(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "abxy"); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Edit().SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Type(ctx, "z"); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("email"); got != "z" {
		t.Fatalf("replaced = %q", got)
	}
}

func TestLogoutClearsFieldsAndHistory(t *testing.T) {
	t.Parallel()

	ownframe.UseMemory(true)
	ownframe.Write("secret")
	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	app.Logout()

	if got := app.Page().Form().FormValue("email"); got != "" {
		t.Fatalf("email = %q", got)
	}

	if got := app.Page().Form().FormValue("password"); got != "" {
		t.Fatalf("password = %q", got)
	}

	if got, _ := ownframe.Read(); got != "" {
		t.Fatalf("clipboard = %q", got)
	}

	if err := app.Page().Edit().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("password"); got != "" {
		t.Fatalf("undo after logout = %q", got)
	}
}

func TestLogoutClearsView(t *testing.T) {
	t.Parallel()

	ownframe.UseMemory(true)
	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "", "login")
	app.Logout()

	if got := app.View().Status; got != "" {
		t.Fatalf("status = %q", got)
	}

	if got := app.View().Error; got != "" {
		t.Fatalf("error = %q", got)
	}

	if app.View().Ready {
		t.Fatal("ready after logout")
	}
}
