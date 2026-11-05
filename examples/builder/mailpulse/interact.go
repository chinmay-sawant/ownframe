package main

// selectableRow returns the row class for a selectable MailPulse row,
// appending the selected modifier when the row is active.
func selectableRow(selected bool) string {
	if selected {
		return "mp-selectable mp-selected"
	}
	return "mp-selectable"
}
