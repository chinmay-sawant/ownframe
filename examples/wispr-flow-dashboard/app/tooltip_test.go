package app

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestInfoTooltipsNeedHover checks every info circle stays hidden until its
// tip is hovered, then shows its bubble text.
func TestInfoTooltipsNeedHover(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := context.Background()

	openInsights(t, app)

	tips := map[string]string{
		"tip-wpm":        "Your average dictation speed this month.",
		"tip-corrected":  "Words Flow corrected as you dictated.",
		"tip-dictionary": "Fixes Flow applied from your personal dictionary.",
	}

	for id, text := range tips {
		tip, ok := findBox(app.Boxes(), id)
		if !ok {
			t.Fatalf("no %s box", id)
		}

		if hasText(app.Boxes(), text) {
			t.Fatalf("%s bubble visible before hover", id)
		}

		if err := app.Page().Hover(ctx, tip.X+tip.W/2, tip.Y+tip.H/2); err != nil {
			t.Fatal(err)
		}

		if !hasText(app.Boxes(), text) {
			t.Fatalf("%s bubble missing on hover", id)
		}
	}
}

func hasText(boxes []ownframe.Box, want string) bool {
	for _, box := range boxes {
		if box.ID == "" && box.Text == want {
			return true
		}
	}

	return false
}
