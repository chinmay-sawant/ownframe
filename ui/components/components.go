// Package components provides experimental themed controls built with ui.
// Constructors return fresh mutable trees. CSS variables configure appearance;
// application state and events stay in Go.
package components

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
type ButtonOptions struct {
	Overrides theme.Overrides
	Icon      *ui.Element
	Variant   ButtonVariant
	Size      ButtonSize
	Disabled  bool
}

// Button builds an HTML button. Filled, outlined, tonal, text and danger
// variants share the global theme; local overrides take precedence.
func Button(label string, options ButtonOptions) *ui.Element {
	bg, fg, border := theme.Control(), theme.Text(), theme.BorderColor()
	switch options.Variant {
	case ButtonFilled:
		bg, fg, border = theme.Primary(), theme.OnPrimary(), theme.Primary()
	case ButtonOutlined:
		bg, fg, border = ui.Transparent(), theme.Primary(), theme.Primary()
	case ButtonText:
		bg, fg, border = ui.Transparent(), theme.Primary(), ui.Transparent()
	case ButtonDanger:
		bg, fg, border = theme.Danger(), theme.OnDanger(), theme.Danger()
	}
	x, y, font := 12.0, 9.0, 12.0
	switch options.Size {
	case ButtonSmall:
		x, y, font = 9, 6, 11
	case ButtonLarge:
		x, y, font = 18, 12, 14
	}
	button := ui.Button(label).Attr("type", "button").BoxSizing(ui.BoxSizingBorderBox).
		BackgroundColor(bg).Color(fg).Border(ui.SolidBorder(ui.Px(1), border)).
		PaddingTop(ui.LengthVar("--of-button-padding-y", ui.Px(y))).
		PaddingBottom(ui.LengthVar("--of-button-padding-y", ui.Px(y))).
		PaddingLeft(ui.LengthVar("--of-button-padding-x", ui.Px(x))).
		PaddingRight(ui.LengthVar("--of-button-padding-x", ui.Px(x))).
		BorderRadius(theme.Radius()).FontSize(ui.Px(font)).FontWeight(ui.FontWeightBold)
	if options.Icon != nil {
		button.Text("").Display(ui.DisplayInlineFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(6)).Children(options.Icon, ui.Text(label))
	}
	if options.Disabled {
		button.Attr("disabled", "disabled").Opacity(ui.Number(disabledOpacity))
	}
	return options.Overrides.Apply(button)
}

const disabledOpacity = 0.5

// Card groups content on the theme surface.
func Card(children ...*ui.Element) *ui.Element {
	return ui.Section().BackgroundColor(theme.Surface()).
		Color(theme.Text()).
		Border(theme.Border()).
		BorderRadius(theme.Radius()).
		Padding(theme.Spacing()).Children(children...)
}

// Badge is a non-interactive text label.
func Badge(label string) *ui.Element {
	return ui.Span().Text(label).BackgroundColor(theme.Selected()).
		Color(theme.Primary()).
		Padding(ui.Edges(ui.Px(4), ui.Px(8))).BorderRadius(theme.Radius()).
		FontSize(ui.Px(11)).FontWeight(ui.FontWeightBold)
}

// TextField creates a labeled text input. Keep id stable across view updates.
// Live user edits are preserved; use Page.Form to set a value after mounting.
func TextField(id, label, value string) *ui.Element {
	return ui.Div().Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(4)).Children(
		ui.Tag("label").Attr("for", id).Text(label).Color(theme.Text()),
		ui.Input().ID(id).Attr("type", "text").Attr("value", value).
			BoxSizing(ui.BoxSizingBorderBox).Width(ui.Percent(100)).
			BackgroundColor(theme.Surface()).Color(theme.Text()).
			Border(theme.Border()).
			Padding(ui.Px(8)).BorderRadius(theme.Radius()),
	)
}
