package main

// The click state of a card lives in the parsed tree. The template renders
// the base markup, and the app patches the state into the tree, so a click
// paints one card instead of rebuilding the page.

// cardClass returns the class list a card carries for the app's state.
func (d *demo) cardClass(row catalogRow) string {
	class := row.BaseClass
	if d.v.Started[row.Button] {
		class += " is-started"
	}
	if d.v.Selected == row.Name {
		class += " is-selected"
	}
	return class
}

// spanClass returns the class list the state span carries.
func (d *demo) spanClass(row catalogRow) string {
	class := "catalog-state"
	if d.v.Started[row.Button] {
		class += " is-started"
	}
	if d.v.Selected == row.Name {
		class += " is-selected"
	}
	return class
}

// spanText returns the text the state span carries.
func (d *demo) spanText(row catalogRow) string {
	if d.v.Started[row.Button] {
		return "Started / selected"
	}
	return "Ready"
}
