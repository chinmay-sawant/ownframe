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
type ButtonOptions struct {
	Overrides theme.Overrides
	Icon      *ui.Element
	Variant   ButtonVariant
	Size      ButtonSize
	Disabled  bool
	Loading   bool
}

// Button builds an HTML button with a label, an optional leading icon, and a
// loading state. Paint and interactive states live in Stylesheet; the host
// adds ID, OnClick, or data-action on the returned element.
func Button(label string, options ButtonOptions) *ui.Element {
	iconOnly := options.Icon != nil && label == ""
	button := ui.Button("").Attr("type", "button").
		Class(classes("of-button", buttonVariantClass(options.Variant), buttonSizeClass(options.Size), buttonIconClass(iconOnly)))
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
		button.Children(ui.Span().Class("of-button-label").Text(label))
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
