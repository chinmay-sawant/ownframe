package login_test

import (
	"context"
	"testing"
)

func TestEmptyLoginDoesNothing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "login")

	if got := app.View().Error; got != "" {
		t.Fatalf("error = %q", got)
	}

	if got := app.View().Status; got != "" {
		t.Fatalf("status = %q", got)
	}

	if got := boxText(t, app, "message"); got != "" {
		t.Fatalf("message = %q", got)
	}
}

func TestSubmitIgnoresEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	if err := app.Submit(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Error; got != "" {
		t.Fatalf("error = %q", got)
	}

	if got := app.View().Status; got != "" {
		t.Fatalf("status = %q", got)
	}
}

func TestWrongLoginThenTypeClearsError(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "wrong"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "wrong"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "", "login")
	if got := app.View().Error; got != "Unknown email or password." {
		t.Fatalf("error before type = %q", got)
	}

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "x"); err != nil {
		t.Fatal(err)
	}

	if got := app.View().Error; got != "" {
		t.Fatalf("error after type = %q", got)
	}

	if err := app.Page().Edit().SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "password", "")
	if err := app.Page().Edit().SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	if got := boxText(t, app, "email"); got != "secret" {
		t.Fatalf("email text = %q", got)
	}

	if got := boxText(t, app, "password"); got != "••••••" {
		t.Fatalf("password text = %q", got)
	}

	click(t, ctx, app, "", "login")

	if got := app.View().Error; got != "" {
		t.Fatalf("error after login = %q", got)
	}

	if got := app.View().Status; got != "Signed in." {
		t.Fatalf("status = %q", got)
	}

	if got := boxText(t, app, "message"); got != "Signed in." {
		t.Fatalf("message = %q", got)
	}
}

func TestPasswordCopyAndDeleteWord(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	text, ok, err := app.Page().Edit().Copy(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if !ok || text != "secret" {
		t.Fatalf("password copy = %q ok=%v", text, ok)
	}

	if err := app.Page().Edit().DeleteWord(ctx); err != nil {
		t.Fatal(err)
	}

	if got := app.Page().Form().FormValue("password"); got != "" {
		t.Fatalf("password after delete word = %q", got)
	}
}

func TestLoginButtonIsAButton(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	login := boxByID(t, app, "login")
	if login.Tag != "button" || login.W <= 0 || login.H <= 0 {
		t.Fatalf("login control = %+v", login)
	}
}

func TestButtonGrayUntilBothFields(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	if got := buttonFill(t, app); got != disabledButton {
		t.Fatalf("empty fill = %+v, want %+v", got, disabledButton)
	}

	click(t, ctx, app, "email", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	if app.View().Ready {
		t.Fatal("ready after email alone")
	}

	if got := buttonFill(t, app); got != disabledButton {
		t.Fatalf("email-only fill = %+v, want %+v", got, disabledButton)
	}

	click(t, ctx, app, "password", "")
	if err := app.Type(ctx, "secret"); err != nil {
		t.Fatal(err)
	}

	if !app.View().Ready {
		t.Fatal("not ready after both fields")
	}

	if got := buttonFill(t, app); got != blueButton {
		t.Fatalf("filled fill = %+v, want %+v", got, blueButton)
	}
}
