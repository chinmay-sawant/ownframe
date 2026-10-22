package console_test

import (
	"context"
	"testing"
)

func TestClosedFilterShowsEmptyList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	signIn(t, ctx, app)

	click(t, ctx, app, "filter-closed", "filter-closed")

	if !app.View().Empty {
		t.Fatal("list was not empty")
	}

	if n := len(app.View().Tickets); n != 0 {
		t.Fatalf("Tickets = %d", n)
	}

	boxByID(t, app, "empty")
}

func TestOpenTicketShowsDetail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	signIn(t, ctx, app)

	click(t, ctx, app, "t1", "open-t1")

	if got := app.View().Screen; got != "detail" {
		t.Fatalf("Screen = %q", got)
	}

	if got := app.View().Current.ID; got != "t1" {
		t.Fatalf("Current.ID = %q", got)
	}

	if got := app.View().Current.Title; got != "Seal mismatch" {
		t.Fatalf("Title = %q", got)
	}

	boxByID(t, app, "detail")
}

func TestListWindowsTickets(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	signIn(t, ctx, app)

	n := len(app.View().Tickets)
	if n == 0 {
		t.Fatal("Tickets window empty")
	}

	if n >= 80 {
		t.Fatalf("Tickets window = %d, want < 80", n)
	}

	boxByID(t, app, "t1")
}
