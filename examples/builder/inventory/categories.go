package main

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// categoryPanel shows inventory value split across categories as a stacked bar
// with a legend. Widths come from sampleCats shares in data.go.
func (a *app) categoryPanel() *ui.Element {
	colors := []string{"#1d5c3f", "#3d8b5f", "#7ab894", "#a9d3bb", "#cfe7d8", "#e3efe7"}

	segments := make([]*ui.Element, 0, len(sampleCats))
	for i, c := range sampleCats {
		color := colors[i%len(colors)]
		segments = append(segments, ui.Div().Class("inv-catseg").
			FlexGrow(ui.Number(c.Share)).
			BackgroundColor(ui.Hex(color)).
			Attr("aria-label", fmt.Sprintf("%s %s (%s)", c.Label, c.Amount, c.Pct)))
	}

	legend := make([]*ui.Element, 0, len(sampleCats))
	for i, c := range sampleCats {
		color := colors[i%len(colors)]
		legend = append(legend, ui.Div().Class("inv-row").Gap(ui.Px(6)).Children(
			ui.Span().Class("inv-legend-swatch").BackgroundColor(ui.Hex(color)),
			ui.Text(c.Label),
			ui.Text(c.Amount).FontWeight(ui.FontWeightBold),
			ui.Text(c.Pct).Class("inv-muted"),
		))
	}

	return ui.Div().Class("inv-card").Children(chadcn.Card(
		ui.Div().Class("inv-card-head").Children(
			ui.Div().Children(
				chadcn.CardTitle("Inventory value by category"),
				ui.Div().Class("inv-card-sub").Text("at landed cost"),
			),
			ui.Text("$2,410,000").FontWeight(ui.FontWeightBold).FontSize(ui.Px(18)),
		),
		ui.Div().Class("inv-catbar").Children(segments...),
		ui.Div().Class("inv-cat-legend").Children(legend...),
	).Gap(ui.Px(12)))
}
