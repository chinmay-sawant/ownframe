package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// SwitchOptions configures a switch beyond its label and checked state.
type SwitchOptions struct {
	// ID is the stable host ID. The track and thumb derive their own IDs
	// from it, so an off->on rebuild patches instead of reparsing.
	ID string
	// Disabled marks the control as non-interactive.
	Disabled bool
}

// Switch is a controlled switch. The host changes checked after activation.
// Assign an ID and OnClick or data-action through the returned builder.
func Switch(label string, checked bool) *ui.Element {
	return SwitchWithOptions(label, checked, SwitchOptions{})
}

// SwitchWithOptions builds a Switch with a disabled state.
func SwitchWithOptions(label string, checked bool, options SwitchOptions) *ui.Element {
	trackColor, thumbColor, offset := theme.Control(), ui.ColorVar(switchThumbProperty, ui.Hex("#ffffff")), switchThumbInset
	if checked {
		trackColor = theme.Primary()
		thumbColor = theme.OnPrimary()
		offset = switchThumbChecked
	}
	track := ui.Div().Class("of-switch-track").Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).
		Width(ui.Px(switchTrackWidth)).Height(ui.Px(switchTrackHeight)).FlexShrink(ui.Number(0)).
		BorderRadius(ui.Px(switchTrackHeight / 2)).BackgroundColor(trackColor)
	thumb := ui.Div().Class("of-switch-thumb").Width(ui.Px(switchThumbSize)).Height(ui.Px(switchThumbSize)).
		FlexShrink(ui.Number(0)).BorderRadius(ui.Px(switchThumbSize / 2)).
		MarginLeft(ui.Px(offset)).BackgroundColor(thumbColor)
	if trackID := partID(options.ID, "-track"); trackID != "" {
		track.ID(trackID)
	}
	if thumbID := partID(options.ID, "-thumb"); thumbID != "" {
		thumb.ID(thumbID)
	}
	track.Children(thumb)
	button := ui.Button("").Attr("type", "button").Class(classes("of-switch")).
		Attr("role", "switch").Attr("aria-checked", ariaBool(checked)).
		Children(track, ui.Span().Class("of-switch-label").Text(label))
	if options.ID != "" {
		button.ID(options.ID)
	}
	return setDisabled(button, options.Disabled)
}

const (
	switchTrackWidth    = 36.0
	switchTrackHeight   = 20.0
	switchThumbSize     = 16.0
	switchThumbInset    = 2.0
	switchThumbChecked  = 18.0
	switchThumbProperty = "--of-switch-thumb"
)
