package chadcn

import (
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// Notice displays a heading and supporting text on a themed inset surface.
func Notice(title, body string) *ui.Element {
	return Panel().BackgroundColor(theme.Selected()).BorderLeft(ui.SolidBorder(ui.Px(3), theme.Primary())).
		Padding(theme.Spacing()).Children(ui.Text(title).FontWeight(ui.FontWeightBold),
		ui.P().Text(body).Margin(ui.Edges(ui.Px(6), ui.Px(0), ui.Px(0))).Color(theme.Text()).FontSize(ui.Px(12)))
}
