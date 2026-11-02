package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// FieldOptions describes one labeled form row. The zero value renders an empty
// wrapper, so a host can start from the default and fill in what it has.
type FieldOptions struct {
	// Label is the visible label text. An empty label is not rendered.
	Label string
	// Description is secondary help text under the control.
	Description string
	// Error is a validation message under the control. A non-empty value
	// marks the wrapper invalid and renders the message as an alert.
	Error string
	// For names the control ID the label points at. Pass it whenever the
	// control has an ID so pointer and screen-reader users reach the control
	// through the label.
	For string
	// Control is the input, textarea, select, or composite the field labels.
	Control *ui.Element
	// Required marks the label with a required indicator.
	Required bool
}

// Field stacks a label, one control, and its messages in a column. The host
// supplies the control, so the field owns no value and needs no lifecycle.
// Rebuild it with View.Update when a description or error changes.
func Field(options FieldOptions) *ui.Element {
	field := ui.Div().Class(classes("of-scope", "of-field", fieldStateClass(options.Error)))
	if options.Label != "" {
		field.Children(fieldLabel(options.For, options.Label, options.Required))
	}
	if options.Control != nil {
		field.Children(options.Control)
	}
	if options.Description != "" {
		field.Children(FieldDescription(options.Description))
	}
	if options.Error != "" {
		field.Children(FieldError(options.Error))
	}
	return field
}

// FieldLabel returns the label part of a field. Chain Attr("for", id) when the
// label points at a control that lives outside Field.
func FieldLabel(text string) *ui.Element {
	return ui.Tag("label").Class("of-field-label").Text(text)
}

// FieldDescription returns the muted help line under a control.
func FieldDescription(text string) *ui.Element {
	return ui.P().Class("of-field-description").Text(text)
}

// FieldError returns a validation message. The role makes assistive
// technology announce it when the host rebuilds the field.
func FieldError(text string) *ui.Element {
	return ui.P().Class("of-field-error").Attr("role", "alert").Text(text)
}

func fieldStateClass(message string) string {
	if message == "" {
		return ""
	}
	return "of-field--invalid"
}

func fieldLabel(forID, text string, required bool) *ui.Element {
	label := FieldLabel(text)
	if forID != "" {
		label.Attr("for", forID)
	}
	if required {
		label.Children(ui.Span().Class("of-field-required").Attr("aria-hidden", "true").Text("*"))
	}
	return label
}
