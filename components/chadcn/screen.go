package chadcn

import (
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
