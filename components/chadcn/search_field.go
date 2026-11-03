package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// SearchField builds an accessible search control. Value initializes the field;
// use Page.Form to change a live value after mounting. The label is the
// accessible name and the placeholder, and the control keeps the input paint
// plus the search surface.
func SearchField(id, label, value string) *ui.Element {
	search := inputElement(id, value, InputOptions{Type: "search", Placeholder: label}, "of-search-field")
	if label != "" {
		search.Attr("aria-label", label)
	}
	return search
}
