package login_test

import (
	"context"
	"testing"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

func TestLogoutClearsFieldsAndHistory(t *testing.T) {
	t.Parallel()

	osclip.UseMemory(true)
	osclip.Write("secret")
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

	if got := app.Page().FormValue("email"); got != "" {
		t.Fatalf("email = %q", got)
	}

	if got := app.Page().FormValue("password"); got != "" {
		t.Fatalf("password = %q", got)
	}

	if got, _ := osclip.Read(); got != "" {
		t.Fatalf("clipboard = %q", got)
	}

	if err := app.Page().Undo(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().FormValue("password"); got != "" {
		t.Fatalf("undo after logout = %q", got)
	}
}

func TestLogoutClearsView(t *testing.T) {
	t.Parallel()

	osclip.UseMemory(true)
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
