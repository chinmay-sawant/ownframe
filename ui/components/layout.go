package components

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// Screen fills the viewport with the global theme's canvas and foreground.
func Screen(children ...*ui.Element) *ui.Element {
	const fullViewport = 100
	return ui.Div().Width(ui.Percent(100)).Height(ui.Vh(fullViewport)).Display(ui.DisplayFlex).
		FlexDirection(ui.FlexDirectionColumn).BackgroundColor(theme.Background()).
		Color(theme.Text()).FontFamily(ui.FontFamilySansSerif).FontSize(ui.Px(14)).Children(children...)
}

// Panel is a themed surface without the card's border or padding.
func Panel(children ...*ui.Element) *ui.Element {
	return ui.Div().BackgroundColor(theme.Surface()).Color(theme.Text()).
		BorderRadius(theme.Radius()).Children(children...)
}

// SearchField builds an accessible search control. Value initializes the field;
// use Page.Form to change a live value after mounting.
func SearchField(id, label, value string) *ui.Element {
	return ui.Input().ID(id).Attr("type", "search").Attr("aria-label", label).
		Attr("placeholder", label).Attr("value", value).BoxSizing(ui.BoxSizingBorderBox).
		Padding(ui.Edges(ui.Px(7), ui.Px(10))).Border(theme.Border()).BorderRadius(theme.Radius()).
		BackgroundColor(theme.Control()).Color(theme.Text())
}

// Choice is a selectable, labeled option with a secondary description.
func Choice(label, caption string, selected bool) *ui.Element {
	background, border := theme.Control(), theme.BorderColor()
	check := ""
	if selected {
		background, border, check = theme.Selected(), theme.Primary(), "✓"
	}
	return ui.Button("").Attr("type", "button").Attr("aria-pressed", strconv.FormatBool(selected)).Width(ui.Percent(100)).TextAlign(ui.TextAlignLeft).
		BackgroundColor(background).Color(theme.Text()).Border(ui.SolidBorder(ui.Px(1), border)).
		BoxSizing(ui.BoxSizingBorderBox).Padding(ui.Edges(ui.Px(12), ui.Px(14))).BorderRadius(theme.Radius()).
		Children(ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Children(
			ui.Div().Children(ui.Text(label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(14)),
				ui.P().Text(caption).Margin(ui.Edges(ui.Px(4), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted())),
			ui.Text(check).FontWeight(ui.FontWeightBold).Color(theme.Primary()),
		))
}
