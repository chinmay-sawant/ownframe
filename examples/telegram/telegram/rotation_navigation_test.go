package telegram_test

import (
	"context"
	"testing"
)

func TestPhoneTabChangeResetsPreviousScroll(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	app.Page().TakeScroll()
	app.Page().SetScrollOffset(0, 320)
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	click(t, ctx, app, "tab-settings", "")
	req, ok := app.Page().TakeScroll()
	if !ok || !req.Absolute || req.X != 0 || req.Y != 0 {
		t.Fatalf("tab change must reset the previous screen offset: %+v, queued=%v", req, ok)
	}
}

func TestPhoneListResizeResetsScroll(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	for _, offset := range []int{10, 20, 320} {
		app.Page().SetSize(885, 420)
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		app.Page().TakeScroll()
		app.Page().SetScrollOffset(0, offset)
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		app.Page().SetSize(420, 934)
		if err := app.Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		req, ok := app.Page().TakeScroll()
		if !ok || !req.Absolute || req.Y != 0 {
			t.Fatalf("list resize kept offset %d: %+v, queued=%v", offset, req, ok)
		}
	}
}

func TestPhoneSearchKeyboardDoesNotJumpToListEnd(t *testing.T) {
	ctx := context.Background()
	app := phoneApp(t)
	app.Page().TakeScroll()
	app.SetInsets(28, 200)
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if req, ok := app.Page().TakeScroll(); ok {
		t.Fatalf("list keyboard queued a thread scroll: %+v", req)
	}
}
