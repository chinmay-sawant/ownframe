package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// TextField creates a labeled text input. Keep id stable across view updates.
// Live user edits are preserved; use Page.Form to set a value after mounting.
// Field builds the same pair with a description, error, and required marker.
func TextField(id, label, value string) *ui.Element {
	return ui.Div().Class(classes("of-scope", "of-field", "of-text-field")).Children(
		fieldLabel(id, label, false),
		Input(id, value),
	)
}
