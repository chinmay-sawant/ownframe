package components

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// Switch is a controlled switch. The host changes checked after activation.
// Assign an ID and OnClick or data-action through the returned builder.
func Switch(label string, checked bool) *ui.Element {
	background := theme.Control()
	offset := 2.0
	if checked {
		background = theme.Primary()
		offset = 18
	}
	return Button("", ButtonOptions{Variant: ButtonText}).Attr("role", "switch").
		Attr("aria-checked", strconv.FormatBool(checked)).Color(theme.Text()).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MinHeight(ui.Px(44)).
		Children(ui.Div().Width(ui.Px(36)).Height(ui.Px(20)).FlexShrink(ui.Number(0)).BorderRadius(ui.Px(10)).
			Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).BackgroundColor(background).
			Children(ui.Div().Width(ui.Px(16)).Height(ui.Px(16)).FlexShrink(ui.Number(0)).BorderRadius(ui.Px(8)).
				MarginLeft(ui.Px(offset)).BackgroundColor(theme.OnPrimary())), ui.Text(label))
}

// Checkbox is a controlled checkbox button with an accessible checked state.
func Checkbox(label string, checked bool) *ui.Element {
	mark := ""
	background := theme.Surface()
	if checked {
		mark = "✓"
		background = theme.Primary()
	}
	return Button("", ButtonOptions{}).Attr("role", "checkbox").Attr("aria-checked", strconv.FormatBool(checked)).
		Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).Gap(ui.Px(10)).MinHeight(ui.Px(44)).TextAlign(ui.TextAlignLeft).
		Children(ui.Div().Text(mark).Display(ui.DisplayFlex).AlignItems(ui.AlignItemsCenter).JustifyContent(ui.JustifyContentCenter).Width(ui.Px(18)).Height(ui.Px(18)).FlexShrink(ui.Number(0)).Border(theme.Border()).
			BorderRadius(ui.Px(4)).BackgroundColor(background).Color(theme.OnPrimary()).
			TextAlign(ui.TextAlignCenter), ui.Text(label))
}

// Progress displays a percentage, clamped to 0..100, with progressbar semantics.
func Progress(label string, value int) *ui.Element {
	const maximum = 100
	value = min(max(value, 0), maximum)
	return ui.Div().Attr("role", "progressbar").Attr("aria-label", label).
		Attr("aria-valuemin", "0").Attr("aria-valuemax", "100").Attr("aria-valuenow", strconv.Itoa(value)).
		Width(ui.Percent(maximum)).Height(ui.Px(8)).Overflow(ui.OverflowHidden).
		BackgroundColor(theme.Control()).BorderRadius(theme.Radius()).
		Children(ui.Div().Width(ui.Percent(float64(value))).Height(ui.Percent(maximum)).
			BackgroundColor(theme.Primary()).BorderRadius(theme.Radius()))
}

// Notice displays a heading and supporting text on a themed inset surface.
func Notice(title, body string) *ui.Element {
	return Panel().BackgroundColor(theme.Selected()).BorderLeft(ui.SolidBorder(ui.Px(3), theme.Primary())).
		Padding(theme.Spacing()).Children(ui.Text(title).FontWeight(ui.FontWeightBold),
		ui.P().Text(body).Margin(ui.Edges(ui.Px(6), ui.Px(0), ui.Px(0))).Color(theme.Text()).FontSize(ui.Px(12)))
}
