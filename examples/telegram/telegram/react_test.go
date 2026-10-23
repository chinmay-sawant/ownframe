package telegram_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

// longPress holds a press at the center of the last box with that id.
func longPress(t *testing.T, ctx context.Context, app *telegram.App, id string) {
	t.Helper()

	x, y := center(t, app, id, "")
	if _, err := app.Page().Testing().LongPress(ctx, x, y); err != nil {
		t.Fatal(err)
	}
}

func TestLongPressOpensReactions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	longPress(t, ctx, app, "msg-anna-5")

	if app.View().ReactID != "anna-5" {
		t.Fatalf("react id = %q", app.View().ReactID)
	}

	boxByID(t, app, "react-like")
	boxByID(t, app, "react-party")
}

func TestClickClosesTheReactionRow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	longPress(t, ctx, app, "msg-anna-5")

	if app.View().ReactID == "" {
		t.Fatal("row did not open")
	}

	click(t, ctx, app, "msg-anna-5", "")

	if app.View().ReactID != "" {
		t.Fatalf("react id = %q", app.View().ReactID)
	}

	if hasBox(app, "react-like") {
		t.Fatal("reaction row still drawn")
	}
}
