package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// ToggleOptions describes the styling of one toggle. The zero value is a
// medium, enabled toggle with no icon.
type ToggleOptions struct {
	Overrides theme.Overrides
	Icon      *ui.Element
	Size      ButtonSize
	Disabled  bool
}

// Toggle builds a two-state button. Pressed is the host's state and becomes
// aria-pressed; CSS paints the pressed face. Toggle reads at most one options
// value, so Toggle("Bold", true) is the default appearance.
func Toggle(label string, pressed bool, options ...ToggleOptions) *ui.Element {
	settings := ToggleOptions{}
	if len(options) > 0 {
		settings = options[0]
	}
	iconOnly := settings.Icon != nil && label == ""
	toggle := ui.Button("").Attr("type", "button").Attr("aria-pressed", ariaBool(pressed)).
		Class(classes("of-toggle", toggleSizeClass(settings.Size), toggleIconClass(iconOnly)))
	if settings.Icon != nil {
		toggle.Children(ui.Span().Class("of-toggle-icon").Attr("aria-hidden", "true").Children(settings.Icon))
	}
	if label != "" {
		toggle.Children(ui.Span().Class("of-toggle-label").Text(label))
	}
	setDisabled(toggle, settings.Disabled)
	return settings.Overrides.Apply(toggle)
}

// toggleSizeClass returns the size modifier class for a toggle.
func toggleSizeClass(size ButtonSize) string {
	switch size {
	case ButtonSmall:
		return "of-toggle--small"
	case ButtonLarge:
		return "of-toggle--large"
	default:
		return "of-toggle--medium"
	}
}

// toggleIconClass marks an icon-only toggle so CSS can size its target.
func toggleIconClass(iconOnly bool) string {
	if iconOnly {
		return "of-toggle--icon"
	}
	return ""
}
