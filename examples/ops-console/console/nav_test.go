package console_test

import (
	"context"
	"testing"
)

func TestSignInOpensList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	signIn(t, ctx, app)

	if got := app.View().Screen; got != "list" {
		t.Fatalf("Screen = %q", got)
	}

	if got := app.View().Operator; got != "dock" {
		t.Fatalf("Operator = %q", got)
	}

	boxByID(t, app, "t1")
	boxByID(t, app, "nav-list")
}

func TestNavClickChangesScreen(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	signIn(t, ctx, app)

	click(t, ctx, app, "nav-report", "nav-report")

	if got := app.View().Screen; got != "report" {
		t.Fatalf("Screen = %q", got)
	}

	boxByID(t, app, "report")

	click(t, ctx, app, "nav-settings", "nav-settings")

	if got := app.View().Screen; got != "settings" {
		t.Fatalf("Screen = %q", got)
	}

	boxByID(t, app, "settings")
}
