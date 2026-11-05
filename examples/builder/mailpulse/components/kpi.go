// Package components holds MailPulse tiles that ChadCN does not ship.
package components

import (
	"strings"

	"github.com/chinmay-sawant/ownframe/ui"
)

// KPICard renders one MailPulse KPI tile. The host owns the data; each call
// constructs a fresh tree. Tint is one of orange, blue, purple, or green and
// selects the card and icon-dot modifiers. Up selects the delta pill tone.
// The delta gains a directional arrow prefix unless it already carries one,
// so data values like "+3.1%" render as "↗ +3.1%" and "-0.2%" as "↘ -0.2%".
func KPICard(label, value, delta, footnote, tint string, up bool) *ui.Element {
	switch tint {
	case "orange", "blue", "purple", "green":
	default:
		tint = "blue"
	}
	deltaClass := "mp-delta mp-delta--down"
	arrow := "↘"
	if up {
		deltaClass = "mp-delta mp-delta--up"
		arrow = "↗"
	}
	text := strings.TrimSpace(delta)
	if text != "" && !strings.HasPrefix(text, "↗") && !strings.HasPrefix(text, "↘") {
		text = arrow + " " + text
	}
	top := ui.Div().Class("mp-kpi-top").Children(
		ui.Span().Class("mp-kpi-dot mp-kpi-dot--"+tint).Text(dotGlyph(tint)),
		ui.Span().Class("mp-kpi-label").Text(label),
		ui.Span().Class("mp-kpi-more").Text("...").CSSRaw("margin-left", "auto"),
	)
	middle := ui.Div().Class("mp-kpi-value-row").
		Display(ui.DisplayFlex).
		AlignItems(ui.AlignItemsCenter).
		Gap(ui.Px(8)).
		Children(
			ui.Tag("strong").Class("mp-kpi-value").Text(value).FontSize(ui.Px(28)).FontWeight(ui.FontWeightBold),
			ui.Span().Class(deltaClass).Text(text),
		)
	foot := ui.Div().Class("mp-kpi-foot").Children(
		ui.Span().Class("mp-kpi-leader").FlexGrow(ui.Number(1)).BorderTop(ui.RawValue("1px dotted #d4d4d2")),
		ui.Span().Class("mp-muted").Text(footnote).TextAlign(ui.TextAlignCenter),
		ui.Span().Class("mp-kpi-leader").FlexGrow(ui.Number(1)).BorderTop(ui.RawValue("1px dotted #d4d4d2")),
	)
	return ui.Div().Class("mp-kpi mp-kpi--"+tint).Children(top, middle, foot)
}

func dotGlyph(tint string) string {
	switch tint {
	case "orange":
		return "✉"
	case "blue":
		return "➤"
	case "purple":
		return "◉"
	case "green":
		return "↩"
	default:
		return "●"
	}
}
