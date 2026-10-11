package telegram_test

import (
	"context"
	"testing"
)

func TestPhoneTabsFollowScrollAndAcceptClicks(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	app.Page().SetSize(885, 420)
	for _, offset := range []int{0, 180, 320} {
		app.Page().SetScrollOffset(0, offset)
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		for _, tab := range []string{"settings", "contacts", "chats"} {
			box := boxByID(t, app, "tab-"+tab)
			if !closeEnough(box.Y-float64(offset), 420-16-60) {
				t.Fatalf("tab %s at scroll %d has viewport Y %.1f", tab, offset, box.Y-float64(offset))
			}
			if err := app.Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			if app.View().Tab != tab {
				t.Fatalf("tap selected %s, want %s", app.View().Tab, tab)
			}
		}
	}
}

func TestPhoneComposerFollowsRotationAndKeyboard(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	click(t, ctx, app, "", "open-anna")
	for _, state := range [][3]int{{420, 934, 16}, {885, 420, 200}, {420, 934, 16}} {
		app.Page().SetSize(state[0], state[1])
		app.SetInsets(28, state[2])
		if err := app.Tick(ctx); err != nil {
			t.Fatal(err)
		}
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		_, offset := app.Page().ScrollOffset()
		box := boxByID(t, app, "composebar")
		if !closeEnough(box.Y+box.H-float64(offset), state[1]-state[2]) {
			t.Fatalf("composer bottom %.1f outside keyboard viewport %v", box.Y+box.H-float64(offset), state)
		}
	}
}
