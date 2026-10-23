package console_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/ops-console/console"
)

func signIn(t *testing.T, ctx context.Context, app *console.App) {
	t.Helper()

	click(t, ctx, app, "name", "")
	if err := app.Type(ctx, "dock"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "pin", "")
	if err := app.Type(ctx, "1947"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "sign-in", "sign-in")
}

func TestWrongPinStaysOnLogin(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "name", "")
	if err := app.Type(ctx, "dock"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "pin", "")
	if err := app.Type(ctx, "0000"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "sign-in", "sign-in")

	if got := app.View().Screen; got != "login" {
		t.Fatalf("Screen = %q", got)
	}

	if got := app.View().Error; got != "Unknown operator or pin." {
		t.Fatalf("Error = %q", got)
	}
}
