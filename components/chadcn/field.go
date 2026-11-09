package chadcn

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// FieldOptions describes one labeled form row. The zero value renders an empty
// wrapper, so a host can start from the default and fill in what it has.
type FieldOptions struct {
	// Label is the visible label text. An empty label is not rendered.
	Label string
	// Description is secondary help text under the control.
	Description string
	// Error is a validation message under the control. A non-empty value
	// marks the wrapper invalid and renders the message as an alert. When
	// Field can see the control, it also sets aria-invalid and
	// aria-describedby on that control.
	Error string
	// For names the control ID the label points at. Pass it whenever the
	// control has an ID so pointer and screen-reader users reach the control
	// through the label. A click on that label focuses the control.
	For string
	// Control is the input, textarea, select, or composite the field labels.
	Control *ui.Element
	// Required marks the label with a required indicator. When Field can see
	// the control, it also sets required and aria-required on that control.
	Required bool
}

// Field stacks a label, one control, and its messages in a column. The host
// supplies the control, so the field owns no value and needs no lifecycle.
// Rebuild it with View.Update when a description or error changes.
func Field(options FieldOptions) *ui.Element {
	field := ui.Div().Class(classes("of-scope", "of-field", fieldStateClass(options.Error)))
	targetID, seen := fieldTarget(options.Control, options.For)
	if seen {
		markFieldControl(options.Control, targetID, options.Required, options.Error != "")
	}
	// Stable part ids let View.Update patch a label, description, or error
	// text edit instead of taking the full render. They derive from the
	// control id, so a field without one renders exactly as before.
	partBase := options.For
	if seen {
		partBase = targetID
	}
	if options.Label != "" {
		label := fieldLabel(options.For, options.Label, options.Required)
		if id := partID(partBase, "-label"); id != "" {
			label.ID(id)
		}
		field.Children(label)
	}
	if options.Control != nil {
		field.Children(options.Control)
	}
	if options.Description != "" {
		description := FieldDescription(options.Description)
		if id := partID(partBase, "-description"); id != "" {
			description.ID(id)
		}
		field.Children(description)
	}
	if options.Error != "" {
		message := FieldError(options.Error)
		if id := partID(partBase, "-error"); id != "" {
			message.ID(id)
		}
		field.Children(message)
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

// fieldTarget reports the id of the control Field can mark. A wrapper whose
// own tag is not the named control is not visible to Field, so the caller
// leaves it alone.
func fieldTarget(control *ui.Element, forID string) (string, bool) {
	if control == nil {
		return "", false
	}
	source, err := ui.Render(control)
	if err != nil {
		return "", false
	}
	tag, raw, ok := openFirst(source)
	if !ok || !fieldControlTag(tag) {
		return "", false
	}
	id := quotedAttr(raw, "id")
	if forID != "" && id != forID {
		return "", false
	}
	if id == "" {
		id = forID
	}
	return id, true
}

func markFieldControl(control *ui.Element, id string, required, invalid bool) {
	if required {
		control.Attr("required", "required")
		control.Attr("aria-required", "true")
	}
	if !invalid {
		return
	}
	control.Attr("aria-invalid", "true")
	if id != "" {
		control.Attr("aria-describedby", partID(id, "-error"))
	}
}

func fieldControlTag(tag string) bool {
	switch tag {
	case "input", "textarea", "select":
		return true
	default:
		return false
	}
}

func openFirst(source string) (string, string, bool) {
	if source == "" || source[0] != '<' {
		return "", "", false
	}
	i := 1
	for i < len(source) && source[i] != ' ' && source[i] != '>' && source[i] != '/' {
		i++
	}
	tag := strings.ToLower(source[1:i])
	end := strings.IndexByte(source[i:], '>')
	if tag == "" || end < 0 {
		return "", "", false
	}
	return tag, source[i : i+end], true
}

func quotedAttr(raw, name string) string {
	needle := name + `="`
	from := 0
	for from < len(raw) {
		at := strings.Index(raw[from:], needle)
		if at < 0 {
			return ""
		}
		at += from
		if at > 0 && raw[at-1] != ' ' {
			from = at + len(needle)
			continue
		}
		rest := raw[at+len(needle):]
		end := strings.IndexByte(rest, '"')
		if end < 0 {
			return ""
		}
		return rest[:end]
	}
	return ""
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
