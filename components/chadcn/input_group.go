package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// InputGroupOptions describes a text input with inline slots.
type InputGroupOptions struct {
	// ID is the input's stable host ID.
	ID string
	// Label is the input's accessible name.
	Label string
	// Type is the input type. The zero value is "text".
	Type string
	// Value initializes the input.
	Value string
	// Placeholder is shown while the input is empty.
	Placeholder string
	// Prefix is an inline slot before the input, such as an icon.
	Prefix *ui.Element
	// Suffix is an inline slot after the input, such as a result count.
	Suffix *ui.Element
	// Addon is a trailing slot separated by a divider, such as a shortcut hint.
	Addon *ui.Element
	// Disabled marks the input as non-interactive.
	Disabled bool
}

// InputGroup renders a bordered wrapper with optional prefix, suffix, and
// addon slots around one text input. The host reads the value from Page.Form.
func InputGroup(options InputGroupOptions) *ui.Element {
	inputType := options.Type
	if inputType == "" {
		inputType = "text"
	}
	input := ui.Input().Class("of-input-group-input").Attr("type", inputType).
		Attr("value", options.Value)
	if options.ID != "" {
		input.ID(options.ID)
	}
	if options.Label != "" {
		input.Attr("aria-label", options.Label)
	}
	if options.Placeholder != "" {
		input.Attr("placeholder", options.Placeholder)
	}
	input = setDisabled(input, options.Disabled)
	group := ui.Div().Class(classes("of-input-group")).Children(
		inputGroupSlot("of-input-group-prefix", options.Prefix),
		input,
		inputGroupSlot("of-input-group-suffix", options.Suffix),
		inputGroupSlot("of-input-group-addon", options.Addon),
	)
	if options.Disabled {
		group.Attr("aria-disabled", ariaBool(true))
	}
	return group
}

// inputGroupSlot wraps one optional slot. It returns nil when unset so the
// caller does not emit an empty element.
func inputGroupSlot(className string, content *ui.Element) *ui.Element {
	if content == nil {
		return nil
	}
	return ui.Span().Class(className).Children(content)
}
