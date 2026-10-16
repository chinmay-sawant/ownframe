package synth

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestClickNotes(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	defer app.Close()

	ctx := context.Background()
	_ = app.onClick(ctx, ownframe.Box{Action: "note-c"})
	if app.view.ActiveNote != "C4" {
		t.Errorf("got note %q, want C4", app.view.ActiveNote)
	}
	if !app.view.BeatLED {
		t.Errorf("expected beat LED to be lit")
	}
	if app.audio.PlayCount != 1 {
		t.Errorf("got play count %d, want 1", app.audio.PlayCount)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "note-cs"})
	if app.view.ActiveNote != "C#4" {
		t.Errorf("got note %q, want C#4", app.view.ActiveNote)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "note-c2"})
	if app.view.ActiveNote != "C5" {
		t.Errorf("got note %q, want C5", app.view.ActiveNote)
	}

	_ = app.onClick(ctx, ownframe.Box{Action: "stop"})
	if app.view.ActiveNote != "--" || app.view.Playing {
		t.Errorf("expected stopped state after stop click")
	}
}
