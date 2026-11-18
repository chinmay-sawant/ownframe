package telegram_test

import (
	"context"
	"testing"
)

// openAnnaNative opens Anna's thread without web mode and returns the full
// chronological history, oldest first.
func openAnnaNative(t *testing.T, ctx context.Context) []string {
	t.Helper()

	native := newApp(t, ctx)
	click(t, ctx, native, "", "open-anna")

	full := native.View().Thread
	if len(full) <= 20 {
		t.Fatalf("native anna thread = %d, want more than one web page", len(full))
	}

	ids := make([]string, len(full))
	for i, m := range full {
		ids[i] = m.ID
	}

	return ids
}

func TestWebThreadShowsLatestWindow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	web := newApp(t, ctx)
	web.SetWebMode(true)
	click(t, ctx, web, "", "open-anna")

	fullIDs := openAnnaNative(t, ctx)

	view := web.View()
	if got := len(view.Thread); got != 20 {
		t.Fatalf("web thread = %d, want 20 (webThreadPage)", got)
	}

	if !view.ThreadMore {
		t.Error("ThreadMore = false, want true when older messages exist")
	}

	if got := view.Thread[0].ID; got != fullIDs[len(fullIDs)-1] {
		t.Errorf("Thread[0] = %q, want newest %q", got, fullIDs[len(fullIDs)-1])
	}

	windowBase := len(fullIDs) - len(view.Thread)
	if got := view.Thread[len(view.Thread)-1].ID; got != fullIDs[windowBase] {
		t.Errorf("oldest in window = %q, want %q", got, fullIDs[windowBase])
	}
}

func TestWebShowEarlierLoadsMore(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	web := newApp(t, ctx)
	web.SetWebMode(true)
	click(t, ctx, web, "", "open-anna")

	fullIDs := openAnnaNative(t, ctx)
	newest := fullIDs[len(fullIDs)-1]

	before := len(web.View().Thread)
	click(t, ctx, web, "show-earlier", "show-earlier")
	after := web.View().Thread

	if len(after) <= before {
		t.Fatalf("thread after show-earlier = %d, want more than %d", len(after), before)
	}

	if len(after) != len(fullIDs) {
		t.Fatalf("thread after show-earlier = %d, want full %d", len(after), len(fullIDs))
	}

	if web.View().ThreadMore {
		t.Error("ThreadMore = true, want false when all messages are loaded")
	}

	if after[0].ID != newest {
		t.Errorf("Thread[0] = %q, want newest %q", after[0].ID, newest)
	}
}

func TestWebSendKeepsNewestFirst(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	app.SetWebMode(true)

	click(t, ctx, app, "", "open-anna")
	click(t, ctx, app, "compose", "")

	if err := app.Type(ctx, "See you at 7!"); err != nil {
		t.Fatal(err)
	}

	if err := app.Submit(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if len(view.Thread) == 0 {
		t.Fatal("web thread is empty after send")
	}

	if got := view.Thread[0].Text; got != "See you at 7!" {
		t.Fatalf("Thread[0] = %q, want sent message", got)
	}

	if !view.Thread[0].Own {
		t.Errorf("Thread[0].Own = false, want true for the sent message")
	}

	if got := len(view.Thread); got != 20 {
		t.Errorf("web thread = %d, want windowed 20 after send", got)
	}
}
