package ipc_test

import (
	"context"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/ipc/ipc"
)

// TestIPCExampleClicks drives the buttons and checks what each one shows.
func TestIPCExampleClicks(t *testing.T) {
	ctx := context.Background()
	app, err := ipc.New()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(app.Cancel)

	if !app.View().Live {
		t.Fatal("Live = false after New, want registrations")
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if len(app.Page().PNG()) == 0 {
		t.Fatal("PNG is empty after Redraw")
	}

	click(t, ctx, app, "send")
	if got := app.View().Received; got != "ping" {
		t.Fatalf("Received = %q, want the default payload ping", got)
	}

	if got := app.View().Sends; got != 1 {
		t.Fatalf("Sends = %d, want 1", got)
	}

	if got := app.View().Status; !strings.Contains(got, "2 listeners") {
		t.Fatalf("Status = %q, want the fan-out count", got)
	}

	click(t, ctx, app, "request")
	if got := app.View().Status; !strings.Contains(got, `"42"`) {
		t.Fatalf("Status = %q, want the doubled 21", got)
	}

	app.Page().SetFormValue("num", "abc")
	click(t, ctx, app, "request")
	if got := app.View().Status; !strings.Contains(got, "not a number") {
		t.Fatalf("Status = %q, want the handler error", got)
	}

	click(t, ctx, app, "missing")
	if got := app.View().Status; !strings.Contains(got, "ErrNoHandler") {
		t.Fatalf("Status = %q, want ErrNoHandler", got)
	}
}
