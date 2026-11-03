package chadcn

import "github.com/chinmay-sawant/ownframe/ui"

// InputSize selects control height, padding, and type scale. The zero value is
// medium.
type InputSize uint8

const (
	InputMedium InputSize = iota
	InputSmall
	InputLarge
)

// InputOptions describes appearance and state for one single-line field. The
// zero value is an enabled medium text input, which is what Input builds.
type InputOptions struct {
	// Type is the HTML input type, such as "text", "email", "password",
	// "search", or "file". An empty value means "text".
	Type string
	// Placeholder is shown while the value is empty.
	Placeholder string
	// Size selects the control height and type scale.
	Size InputSize
	// Disabled prevents edits and focus.
	Disabled bool
	// Invalid marks the field with aria-invalid so CSS can paint the error
	// border and assistive technology reports the state.
	Invalid bool
}

// Input builds a single-line text field with an initial value. The host owns
// the live value through Page.Form; keep id stable across view updates and let
// the page carry user edits.
func Input(id, value string) *ui.Element {
	return InputWithOptions(id, value, InputOptions{})
}

// InputWithOptions builds an Input with a type, size, placeholder, and state.
func InputWithOptions(id, value string, options InputOptions) *ui.Element {
	return inputElement(id, value, options, "")
}

// inputElement is the shared constructor. extra carries a derived class for a
// component that reuses the input paint, such as SearchField.
func inputElement(id, value string, options InputOptions, extra string) *ui.Element {
	kind := options.Type
	if kind == "" {
		kind = "text"
	}
	input := ui.Input().ID(id).Attr("type", kind).Attr("value", value).
		Class(classes("of-scope", "of-input", "of-focus", inputSizeClass(options.Size), extra))
	if options.Placeholder != "" {
		input.Attr("placeholder", options.Placeholder)
	}
	if options.Invalid {
		input.Attr("aria-invalid", ariaBool(true))
	}
	return setDisabled(input, options.Disabled)
}

func inputSizeClass(size InputSize) string {
	switch size {
	case InputSmall:
		return "of-input--sm"
	case InputLarge:
		return "of-input--lg"
	default:
		return ""
	}
}
