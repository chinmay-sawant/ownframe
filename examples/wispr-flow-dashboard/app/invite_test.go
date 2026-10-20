package app

import (
	"context"
	"testing"
)

// TestInvitePage checks the invite root, the link controls, both lists, and
// the display-list replay path.
func TestInvitePage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("invite")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if _, ok := findBox(app.Boxes(), "page-invite"); !ok {
		t.Fatal("no invite page")
	}

	if app.Page().Display() == nil {
		t.Fatal("invite fell back to the bitmap path")
	}

	view := app.View().Invite

	if view.URL != inviteURL {
		t.Fatalf("URL = %q", view.URL)
	}

	if len(view.Pending) != 2 || len(view.Members) != 4 {
		t.Fatalf("pending = %d members = %d", len(view.Pending), len(view.Members))
	}

	for _, id := range []string{"inv-url", "inv-copy", "inv-email", "inv-send"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}

	copyLink, _ := findBox(app.Boxes(), "inv-copy")

	clickBox(t, app, copyLink)

	if app.View().Note != "Invite link copied" {
		t.Fatalf("Note = %q", app.View().Note)
	}
}
