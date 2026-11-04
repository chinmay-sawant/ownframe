package main

import "context"

// selectableRow returns the row class for a selectable MailPulse row,
// appending the selected modifier when the row is active.
func selectableRow(selected bool) string {
	if selected {
		return "mp-selectable mp-selected"
	}
	return "mp-selectable"
}

// toastSelect shows a selection toast and refreshes the view.
// Core agent wires selection state; this helper only uses existing fields.
func (a *app) toastSelect(ctx context.Context, label string) error {
	a.toast = "Selected " + label
	a.toastOpen = true
	return a.refresh(ctx)
}
