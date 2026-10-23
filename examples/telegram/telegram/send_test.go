package telegram_test

import (
	"context"
	"testing"
)

func TestSendFromComposer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	click(t, ctx, app, "compose", "")

	base := len(app.View().Thread)

	if err := app.Type(ctx, "See you soon!"); err != nil {
		t.Fatal(err)
	}

	if err := app.Submit(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if got := len(view.Thread); got != base+1 {
		t.Fatalf("thread = %d", got)
	}

	last := view.Thread[len(view.Thread)-1]
	if last.Text != "See you soon!" || !last.Own || !last.Read {
		t.Fatalf("last = %+v", last)
	}

	if view.Draft != "" {
		t.Fatalf("draft = %q", view.Draft)
	}

	if got := app.Page().Form().FormValue("compose"); got != "" {
		t.Fatalf("compose value = %q", got)
	}

	for _, c := range view.Chats {
		if c.ID == "anna" && c.Preview != "You: See you soon!" {
			t.Fatalf("preview = %q", c.Preview)
		}
	}
}

func TestSendButton(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	click(t, ctx, app, "compose", "")

	base := len(app.View().Thread)

	if err := app.Type(ctx, "Hi"); err != nil {
		t.Fatal(err)
	}

	click(t, ctx, app, "", "send")

	if got := len(app.View().Thread); got != base+1 {
		t.Fatalf("thread = %d", got)
	}
}

func TestSendIgnoresAnEmptyDraft(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)

	click(t, ctx, app, "", "open-anna")
	base := len(app.View().Thread)
	click(t, ctx, app, "", "send")

	if got := len(app.View().Thread); got != base {
		t.Fatalf("thread = %d", got)
	}
}
