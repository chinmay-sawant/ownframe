package app

import (
	"context"
	"testing"
)

// TestVoicePage checks the voice root, the record button, and the sample rows.
func TestVoicePage(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	app.show("voice")

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	for _, id := range []string{"page-voice", "voice-record", "voice-samples", "voice-details"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("the voice page fell back to the bitmap path")
	}

	voice := DefaultView().Voice

	if len(voice.Steps) != 3 {
		t.Fatalf("steps = %d", len(voice.Steps))
	}

	if len(voice.Details) != 3 {
		t.Fatalf("details = %d", len(voice.Details))
	}

	status := voice.Details[2]

	if !status.Tag || status.Value != "Not set up" {
		t.Fatalf("status detail = %+v", status)
	}
}
