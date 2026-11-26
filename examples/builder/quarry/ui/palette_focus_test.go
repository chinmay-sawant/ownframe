package ui

import "testing"

// TestPaletteInputTakesFocusOnOpen: opening the palette focuses its search
// field, so typing filters at once without a click.
func TestPaletteInputTakesFocusOnOpen(t *testing.T) {
	r := newRig(t, 1440, 960, "")
	r.chord("k", false)
	r.must(r.a.Page().Redraw(r.ctx))
	if got := r.a.Page().Form().FocusID(); got != "pal-input" {
		t.Fatalf("focus = %q, want pal-input", got)
	}
}
