package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// CheckboxOptions configures a checkbox beyond its label and checked state.
type CheckboxOptions struct {
	// ID is the stable host ID. The check mark derives its own ID from it,
	// so an unchecked->checked rebuild patches instead of reparsing.
	ID string
	// Description is secondary text shown below the label.
	Description string
	// Disabled marks the control as non-interactive.
	Disabled bool
}

// Checkbox is a controlled checkbox button with an accessible checked state.
// The host flips checked and rebuilds the tree.
func Checkbox(label string, checked bool) *ui.Element {
	return CheckboxWithOptions(label, checked, CheckboxOptions{})
}

// CheckboxWithOptions builds a Checkbox with a description and disabled state.
func CheckboxWithOptions(label string, checked bool, options CheckboxOptions) *ui.Element {
	// The mark is always rendered, empty when unchecked, so a flip is a
	// text edit on an id-bearing node instead of a structural change.
	markText := ""
	if checked {
		markText = "✓"
	}
	mark := ui.Span().Class("of-checkbox-mark").Text(markText)
	if markID := partID(options.ID, "-mark"); markID != "" {
		mark.ID(markID)
	}
	box := ui.Span().Class("of-checkbox-box").Attr("aria-hidden", "true").Children(mark)
	content := ui.Span().Class("of-checkbox-content").Children(
		ui.Span().Class("of-checkbox-label").Text(label),
	)
	if options.Description != "" {
		content.Children(ui.Span().Class("of-checkbox-description").Text(options.Description))
	}
	checkedClass := ""
	if checked {
		checkedClass = "of-checkbox--checked"
	}
	button := ui.Button("").Attr("type", "button").Class(classes("of-checkbox", checkedClass)).
		Attr("role", "checkbox").Attr("aria-checked", ariaBool(checked)).
		Children(box, content)
	if options.ID != "" {
		button.ID(options.ID)
	}
	return setDisabled(button, options.Disabled)
}
