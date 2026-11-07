package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// Choice is a selectable, labeled option with a secondary description.
func Choice(label, caption string, selected bool) *ui.Element {
	return ChoiceWithID("", label, caption, selected)
}

// ChoiceWithID builds a Choice with id as the root id. The inner parts
// derive stable ids from it: the label span takes id+"-label", the caption
// takes id+"-caption", and the check mark takes id+"-check". The label span
// also takes display:block so it owns a block box: the engine only patches
// an element that owns one, so a fitting label text edit patches instead of
// taking the full render. An empty id renders exactly like Choice, so hosts
// that do not need patches pass nothing.
func ChoiceWithID(id, label, caption string, selected bool) *ui.Element {
	background, border := theme.Control(), theme.BorderColor()
	check := ""
	if selected {
		background, border, check = theme.Selected(), theme.Primary(), "✓"
	}
	labelText := ui.Text(label).FontWeight(ui.FontWeightBold).FontSize(ui.Px(14))
	if part := partID(id, "-label"); part != "" {
		labelText.Display(ui.DisplayBlock).ID(part)
	}
	captionText := ui.P().Text(caption).Margin(ui.Edges(ui.Px(4), ui.Px(0), ui.Px(0))).FontSize(ui.Px(12)).Color(theme.Muted())
	if part := partID(id, "-caption"); part != "" {
		captionText.ID(part)
	}
	checkText := ui.Text(check).FontWeight(ui.FontWeightBold).Color(theme.Primary())
	if part := partID(id, "-check"); part != "" {
		checkText.ID(part)
	}
	root := ui.Button("").Attr("type", "button").Attr("aria-pressed", strconv.FormatBool(selected)).Width(ui.Percent(100)).TextAlign(ui.TextAlignLeft).
		BackgroundColor(background).Color(theme.Text()).Border(ui.SolidBorder(ui.Px(1), border)).
		BoxSizing(ui.BoxSizingBorderBox).Padding(ui.Edges(ui.Px(12), ui.Px(14))).BorderRadius(theme.Radius()).
		Children(ui.Div().Display(ui.DisplayFlex).JustifyContent(ui.JustifyContentSpaceBetween).
			AlignItems(ui.AlignItemsCenter).Children(
			ui.Div().Children(labelText, captionText),
			checkText,
		))
	if id != "" {
		root.ID(id)
	}
	return root
}
