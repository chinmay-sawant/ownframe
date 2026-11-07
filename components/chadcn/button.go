package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// ButtonVariant selects a shared appearance. The zero value is tonal.
type ButtonVariant uint8

const (
	ButtonTonal ButtonVariant = iota
	ButtonFilled
	ButtonOutlined
	ButtonText
	ButtonDanger
)

// ButtonSize selects control padding and typography. The zero value is medium.
type ButtonSize uint8

const (
	ButtonMedium ButtonSize = iota
	ButtonSmall
	ButtonLarge
)

// ButtonOptions describes appearance, state, and local theme properties.
// Loading replaces the icon with a spinner and disables the button.
// ID is the stable host ID: it becomes the button id and the label span
// derives id+"-label" from it, so a label change patches instead of taking
// the full render. Leave it empty when the host does not need patches.
type ButtonOptions struct {
	Overrides theme.Overrides
	Icon      *ui.Element
	ID        string
	Variant   ButtonVariant
	Size      ButtonSize
	Disabled  bool
	Loading   bool
}

// Button builds an HTML button with a label, an optional leading icon, and a
// loading state. Paint and interactive states live in Stylesheet; the host
// adds OnClick or data-action on the returned element, or sets Options.ID
// for a stable button id with a patchable label span.
func Button(label string, options ButtonOptions) *ui.Element {
	iconOnly := options.Icon != nil && label == ""
	button := ui.Button("").Attr("type", "button").
		Class(classes("of-button", buttonVariantClass(options.Variant), buttonSizeClass(options.Size), buttonIconClass(iconOnly)))
	if options.ID != "" {
		button.ID(options.ID)
	}
	if options.Loading {
		button.Attr("aria-busy", "true")
	}
	if options.Icon != nil && !options.Loading {
		button.Children(ui.Span().Class("of-button-icon").Attr("aria-hidden", "true").Children(options.Icon))
	}
	if options.Loading {
		button.Children(ui.Span().Class("of-button-spinner").Attr("aria-hidden", "true"))
	}
	if label != "" {
		labelSpan := ui.Span().Class("of-button-label").Text(label)
		if id := partID(options.ID, "-label"); id != "" {
			labelSpan.ID(id)
		}
		button.Children(labelSpan)
	}
	setDisabled(button, options.Disabled || options.Loading)
	return options.Overrides.Apply(button)
}

func buttonVariantClass(variant ButtonVariant) string {
	switch variant {
	case ButtonFilled:
		return "of-button--filled"
	case ButtonOutlined:
		return "of-button--outlined"
	case ButtonText:
		return "of-button--text"
	case ButtonDanger:
		return "of-button--danger"
	default:
		return "of-button--tonal"
	}
}

// buttonSizeClass returns the size modifier class for a button.
func buttonSizeClass(size ButtonSize) string {
	switch size {
	case ButtonSmall:
		return "of-button--small"
	case ButtonLarge:
		return "of-button--large"
	default:
		return "of-button--medium"
	}
}

// buttonIconClass marks an icon-only button so CSS can square its padding.
func buttonIconClass(iconOnly bool) string {
	if iconOnly {
		return "of-button--icon"
	}
	return ""
}
