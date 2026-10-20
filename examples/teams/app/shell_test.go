package app

import (
	"testing"
)

// TestDefaultSection checks the app opens on Chat with the rail drawn.
func TestDefaultSection(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	if app.View().Section != "chat" {
		t.Fatalf("Section = %q", app.View().Section)
	}

	if _, ok := findBox(app.Boxes(), "page-chat"); !ok {
		t.Fatal("no page-chat box")
	}

	for _, name := range []string{"activity", "chat", "channels", "calendar", "calls", "files"} {
		openSection(t, app, name)

		if app.View().Section != name {
			t.Fatalf("Section = %q after rail-%s", app.View().Section, name)
		}

		if _, ok := findBox(app.Boxes(), "page-"+name); !ok {
			t.Fatalf("no page-%s box", name)
		}
	}
}
