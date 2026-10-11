package telegram_test

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/telegram/telegram"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

func phoneApp(t *testing.T) *telegram.App {
	t.Helper()
	ctx := context.Background()
	app, err := telegram.New()
	if err != nil {
		t.Fatal(err)
	}
	if err := page.PrepareMobile(ctx, app.Page()); err != nil {
		t.Fatal(err)
	}
	app.SetInsets(28, 16)
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	return app
}

func TestPhoneLandscapeKeepsTabsInsideViewport(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	for _, size := range [][2]int{{420, 934}, {885, 420}, {420, 934}} {
		app.Page().SetSize(size[0], size[1])
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		tab := boxByID(t, app, "tab-settings")
		if tab.Y+tab.H > float64(size[1]-16)+1 {
			t.Fatalf("viewport %v: tabs bottom %.1f is outside visible height %d", size, tab.Y+tab.H, size[1]-16)
		}
	}
}
