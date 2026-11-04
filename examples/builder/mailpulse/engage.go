package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const engageDotCount = 10

func (a *app) engageCard() *ui.Element {
	rows := engageRowsFor(a.rangeDays())
	icons := []string{"▲", "●", "◆", "■", "★"}
	cards := make([]*ui.Element, 0, len(rows))
	for i, row := range rows {
		icon := ""
		if i < len(icons) {
			icon = icons[i]
		}
		cards = append(cards, engageRowCard(icon, row))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("What readers engage with"),
			chadcn.CardDescription("Average open and click rate by issue type, "+a.rangeWindow()+"."),
		),
		chadcn.CardContent(cards...),
	)
}

func engageRowCard(icon string, row engageRow) *ui.Element {
	dots := make([]*ui.Element, 0, engageDotCount)
	for i := 0; i < engageDotCount; i++ {
		class := "mp-dot"
		if i < row.filled {
			class += " mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-linkrow").Children(
		ui.Span().Class("mp-engage-icon").Text(icon),
		ui.Div().Class("mp-link-meta").Children(
			ui.Span().Class("mp-engage-label").Text(row.label),
			ui.Div().Class("mp-dots").Children(dots...),
		),
		ui.Div().Class("mp-engage-stats").Children(
			ui.Span().Class("mp-engage-open").Text(row.openPct).FontWeight(ui.FontWeightBold),
			ui.Span().Class("mp-muted").Text(row.clickPct),
		),
	)
}
