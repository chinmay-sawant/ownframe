package telegram_test

import (
	"context"
	"testing"
)

func TestKeyboardBackQueuesBlurAndPreservesDraft(t *testing.T) {
	ctx := context.Background()
	app := newApp(t, ctx)
	click(t, ctx, app, "", "open-anna")
	id := "compose"
	if err := app.Page().Focus(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Type(ctx, "draft"); err != nil {
		t.Fatal(err)
	}
	app.RequestBlur()
	if app.Page().FocusedField() == "" {
		t.Fatal("UI thread changed focus")
	}
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}
	if app.Page().FocusedField() != "" {
		t.Fatal("focus survived keyboard Back")
	}
	if value := app.Page().FormValue(id); value != "draft" {
		t.Fatalf("draft %q", value)
	}
	if app.View().Active == "" {
		t.Fatal("keyboard Back navigated out of chat")
	}
}
