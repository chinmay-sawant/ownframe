package app

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/wispr-flow-dashboard/dictation"
)

// TestDictationPageDraws checks the landing page renders its boxes on the
// display-list path.
func TestDictationPageDraws(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	for _, id := range []string{"page-dictation", "dict-start", "dict-search"} {
		if _, ok := findBox(app.Boxes(), id); !ok {
			t.Fatalf("no %s box", id)
		}
	}

	voice, ok := findBox(app.Boxes(), "voice-profile")
	if !ok {
		t.Fatal("no voice-profile box")
	}

	if voice.Action != "nav-voice" {
		t.Fatalf("voice-profile action = %q", voice.Action)
	}

	if app.Page().Probe().Display() == nil {
		t.Fatal("dictation fell back to the bitmap path")
	}
}

// TestDictationRows checks the six sample rows.
func TestDictationRows(t *testing.T) {
	d := dictation.Default()

	if d.User != "Chinmay" || len(d.Rows) != 6 {
		t.Fatalf("User = %q, len(Rows) = %d", d.User, len(d.Rows))
	}

	for i, row := range d.Rows {
		if row.Time == "" || row.Text == "" {
			t.Fatalf("row %d = %+v", i, row)
		}
	}
}

// TestVoiceProfileClick checks the whole panel opens the voice page.
func TestVoiceProfileClick(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if err := app.Page().RedrawIfChanged(context.Background()); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	voice, ok := findBox(app.Boxes(), "voice-profile")
	if !ok {
		t.Fatal("no voice-profile box")
	}

	clickBox(t, app, voice)

	if got := app.View().ActivePage; got != "voice" {
		t.Fatalf("ActivePage = %q", got)
	}
}
