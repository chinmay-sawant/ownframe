package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/examples/builder/mailpulse/components"
	"github.com/chinmay-sawant/ownframe/ui"
)

const funnelDotTotal = 20

func (a *app) funnelCard() *ui.Element {
	// funnelRows arrives in descending pct order
	// (100.0 > 47.8 > 18.6 > 2.4 > 0.2); keep that order so the
	// pct column reads top-to-bottom in sequence.
	rows := funnelRowsFor(a.rangeDays())
	items := make([]*ui.Element, 0, len(rows)+1)
	for _, row := range rows {
		items = append(items, ui.Div().Class("mp-funnel-row").Children(
			ui.Div().Class("mp-funnel-top").Children(
				ui.Span().Class("mp-funnel-label").Text(row.label),
				ui.Div().Class("mp-funnel-stats").Children(
					ui.Span().Class("mp-funnel-count").Text(row.count),
					ui.Span().Class("mp-funnel-pct").Text(row.pct),
				),
			),
			components.DotBar(row.filled, funnelDotTotal),
		))
	}
	items = append(items, ui.Div().Class("mp-insight").Text("✦ Issues with 3 links or fewer get 22% more clicks."))
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Engagement funnel"),
			chadcn.CardDescription("Average per issue, "+a.rangeWindow()+"."),
		),
		chadcn.CardContent(items...),
	)
}
