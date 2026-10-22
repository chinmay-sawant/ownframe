package telegram_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
)

// replay checks the page still draws from its display list. A false gate
// result drops the window to the bitmap fallback and its badge.
func replay(t *testing.T, ctx context.Context, app *telegram.App) {
	t.Helper()

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if app.Page().Display() == nil {
		t.Fatal("page fell back to the bitmap")
	}
}

// TestThreadsStayReplayable covers every chat, the attachment sheet, and the
// tabs. The design thread regressed once: its wrapped bubbles let the flex
// pass stretch the avatar circles into ellipses, which the replay gate
// rejects.
func TestThreadsStayReplayable(t *testing.T) {
	ctx := context.Background()

	for _, id := range []string{"anna", "design", "devs", "max", "mom", "sofia", "alex"} {
		t.Run(id, func(t *testing.T) {
			app := newApp(t, ctx)
			click(t, ctx, app, "", "open-"+id)
			replay(t, ctx, app)
		})
	}

	t.Run("design-attach", func(t *testing.T) {
		app := newApp(t, ctx)
		click(t, ctx, app, "", "open-design")
		click(t, ctx, app, "attach", "")
		replay(t, ctx, app)
	})

	t.Run("design-scrolled", func(t *testing.T) {
		app := newApp(t, ctx)
		click(t, ctx, app, "", "open-design")
		app.Page().SetScrollOffset(0, 600)
		replay(t, ctx, app)
	})

	t.Run("tabs", func(t *testing.T) {
		app := newApp(t, ctx)
		click(t, ctx, app, "", "tab-contacts")
		replay(t, ctx, app)
		click(t, ctx, app, "", "tab-settings")
		replay(t, ctx, app)
		click(t, ctx, app, "dark", "")
		replay(t, ctx, app)
	})
}
