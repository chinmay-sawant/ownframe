package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

func (a *app) linksCard() *ui.Element {
	days := a.rangeDays()
	quarterWord := "month"
	switch days {
	case 7:
		quarterWord = "week"
	case 90:
		quarterWord = "quarter"
	}
	links := topLinksFor(days)
	rows := make([]*ui.Element, 0, len(links))
	for _, link := range links {
		rows = append(rows, ui.Div().Class("mp-linkrow mp-enter").Children(
			ui.Span().Class("mp-linkicon").Text("↗"),
			ui.Div().Class("mp-linkmain").Children(
				ui.Text(link.url).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
				ui.Text(link.context).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
			),
			ui.Div().Class("mp-linkstats").Children(
				ui.Text(link.clicksText).FontWeight(ui.FontWeightBold),
				ui.Text(link.share).Color(ui.Hex("#64748b")),
			),
		))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Top links this "+quarterWord),
			chadcn.CardDescription("Unique clicks across all sends."),
		),
		chadcn.CardContent(rows...),
	)
}
