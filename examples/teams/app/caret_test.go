package app

import (
	"context"
	"testing"
)

// TestComposeSelectionFollowsTheTheme checks the caret takes the field's
// text color and that a select-all paints the theme highlight, not the
// library's light default, on both themes.
func TestComposeSelectionFollowsTheTheme(t *testing.T) {
	t.Parallel()
	app := newTestApp(t)

	// Dark: caret #f0f0f0, selection #363a5e. Light: #242424 and #e8ebfa.
	checkField(t, app, rgb(0xf0, 0xf0, 0xf0), rgb(0x36, 0x3a, 0x5e))

	clickID(t, app, "me")
	clickID(t, app, "theme-light")
	checkField(t, app, rgb(0x24, 0x24, 0x24), rgb(0xe8, 0xeb, 0xfa))
}

// checkField reads the caret of the focused compose field, then selects all
// and reads the highlight it paints. The first click can land on the scrim
// of an open flyout, so it clicks again until the field holds the focus.
func checkField(t *testing.T, app *App, caret, sel [3]float64) {
	t.Helper()

	for range 2 {
		clickID(t, app, "chat-compose")

		if app.Page().FocusedField() == "chat-compose" {
			break
		}
	}

	if app.Page().FocusedField() != "chat-compose" {
		t.Fatal("the compose field never took the focus")
	}

	if err := app.Page().Testing().Type(context.Background(), "ab"); err != nil {
		t.Fatalf("Type: %v", err)
	}

	gotCaret, _ := fieldColors(app)
	assertColor(t, "caret", gotCaret, caret)

	if err := app.Page().SelectAll(context.Background()); err != nil {
		t.Fatalf("SelectAll: %v", err)
	}

	_, gotSel := fieldColors(app)
	assertColor(t, "selection", gotSel, sel)
}
