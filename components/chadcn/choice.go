package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

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
