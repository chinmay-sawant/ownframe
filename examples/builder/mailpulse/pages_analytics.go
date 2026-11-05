//nolint:mnd // Sample analytics fixtures are literal display content: chart values and dot counts are fixed demo values.
package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

const analyticsDotTotal = 10

type analyticsKPIItem struct {
	label string
	value string
	extra string
}

func analyticsKPIs() []analyticsKPIItem {
	return []analyticsKPIItem{
		{label: "Delivered", value: "212k", extra: ""},
		{label: "Opens", value: "98k", extra: "46.2%"},
		{label: "Clicks", value: "38k", extra: "17.9%"},
		{label: "Unsubs", value: "1.1k", extra: "0.5%"},
	}
}

func analyticsKPICard(item analyticsKPIItem) *ui.Element {
	middle := []*ui.Element{
		ui.Tag("strong").Class("mp-kpi-value").Text(item.value).FontSize(ui.Px(28)).FontWeight(ui.FontWeightBold),
	}
	if item.extra != "" {
		middle = append(middle, ui.Span().Class("mp-delta mp-delta--up").Text(item.extra))
	}
	return ui.Div().Class("mp-kpi").Children(
		ui.Div().Class("mp-kpi-top").Children(
			ui.Span().Class("mp-kpi-label").Text(item.label),
		),
		ui.Div().Class("mp-kpi-value-row").
			Display(ui.DisplayFlex).
			AlignItems(ui.AlignItemsCenter).
			Gap(ui.Px(8)).
			Children(middle...),
		ui.Div().Class("mp-kpi-foot").Children(
			ui.Span().Class("mp-kpi-leader").FlexGrow(ui.Number(1)),
			ui.Span().Class("mp-muted").Text(item.value+" "+item.label).TextAlign(ui.TextAlignCenter),
			ui.Span().Class("mp-kpi-leader").FlexGrow(ui.Number(1)),
		),
	)
}

type weekOpensItem struct {
	week   string
	height int
}

func weekOpens() []weekOpensItem {
	return []weekOpensItem{
		{week: "W1", height: 42},
		{week: "W2", height: 55},
		{week: "W3", height: 48},
		{week: "W4", height: 60},
		{week: "W5", height: 52},
		{week: "W6", height: 68},
		{week: "W7", height: 61},
		{week: "W8", height: 57},
		{week: "W9", height: 70},
		{week: "W10", height: 64},
		{week: "W11", height: 58},
		{week: "W12", height: 72},
	}
}

func analyticsDots(filled int) *ui.Element {
	if filled < 0 {
		filled = 0
	}
	if filled > analyticsDotTotal {
		filled = analyticsDotTotal
	}
	dots := make([]*ui.Element, 0, analyticsDotTotal)
	for i := 0; i < analyticsDotTotal; i++ {
		class := "mp-dot mp-dot--off"
		if i < filled {
			class = "mp-dot mp-dot--on"
		}
		dots = append(dots, ui.Span().Class(class))
	}
	return ui.Div().Class("mp-dots").Children(dots...)
}

type clientClickItem struct {
	label  string
	share  string
	filled int
}

func clientClicks() []clientClickItem {
	return []clientClickItem{
		{label: "Gmail", share: "44%", filled: 4},
		{label: "Apple Mail", share: "31%", filled: 3},
		{label: "Outlook", share: "14%", filled: 1},
		{label: "Other", share: "11%", filled: 1},
	}
}

func (a *app) analyticsPage() *ui.Element {
	kpiItems := analyticsKPIs()
	kpiCards := make([]*ui.Element, 0, len(kpiItems))
	for _, item := range kpiItems {
		kpiCards = append(kpiCards, analyticsKPICard(item))
	}

	weeks := weekOpens()
	bars := make([]*ui.Element, 0, len(weeks))
	for _, w := range weeks {
		h := float64(w.height)
		bars = append(bars, ui.Div().Class("mp-bar-col").
			Flex(ui.Number(1)).
			MinWidth(ui.Px(0)).
			Children(
				ui.Div().Class("mp-bar").
					Height(ui.Px(h)).
					MinHeight(ui.Px(0)).
					MaxWidth(ui.Px(24)),
				ui.Span().Class("mp-muted").Text(w.week).
					FontSize(ui.Px(11)).
					WhiteSpace(ui.WhiteSpaceNowrap),
			))
	}
	opensCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Opens by week"),
			chadcn.CardDescription("Unique opens per send week."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-bars").
				Display(ui.DisplayFlex).
				AlignItems(ui.AlignItemsEnd).
				Gap(ui.Px(4)).
				Overflow(ui.OverflowHidden).
				MinWidth(ui.Px(0)).
				Children(bars...),
		),
	)

	clients := clientClicks()
	rows := make([]*ui.Element, 0, len(clients))
	for _, c := range clients {
		rows = append(rows, ui.Div().Class("mp-linkrow").Children(
			ui.Div().Class("mp-link-meta").Children(
				ui.Span().Class("mp-engage-label").Text(c.label),
				analyticsDots(c.filled),
			),
			ui.Div().Class("mp-engage-stats").Children(
				ui.Span().Class("mp-engage-open").Text(c.share).FontWeight(ui.FontWeightBold),
			),
		))
	}
	clientsCard := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Clicks by client"),
			chadcn.CardDescription("Share of clicks by mail client."),
		),
		chadcn.CardContent(rows...),
	)

	return ui.Div().Class("mp-card mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Analytics"),
				chadcn.Muted("Cross-issue trends."),
			),
		),
		ui.Div().Class("mp-kpis").Children(kpiCards...),
		ui.Div().Class("mp-grid-bot").Children(
			opensCard,
			clientsCard,
		),
	)
}

type reportItem struct {
	id      string
	label   string
	format  string
	cadence string
}

func reportRows() []reportItem {
	return []reportItem{
		{id: "monthly-engagement", label: "Monthly engagement", format: "PDF", cadence: "Weekly"},
		{id: "churn-watch", label: "Churn watch", format: "CSV", cadence: "Monthly"},
		{id: "sponsor-report", label: "Sponsor report", format: "PDF", cadence: "Per send"},
		{id: "growth", label: "Growth", format: "CSV", cadence: "Weekly"},
		{id: "deliverability", label: "Deliverability", format: "PDF", cadence: "Monthly"},
	}
}

func (a *app) reportsPage() *ui.Element {
	items := reportRows()
	rows := make([]*ui.Element, 0, len(items))
	for _, item := range items {
		download := chadcn.Button("Download", chadcn.ButtonOptions{
			Variant: chadcn.ButtonOutlined,
			Size:    chadcn.ButtonSmall,
		}).Attr("data-action", "export").FlexShrink(ui.Number(0))
		rows = append(rows, ui.Div().Class("mp-linkrow mp-report-row "+selectableRow(false)).
			Attr("data-action", "select:report:"+item.id+":"+item.label).
			Children(
				ui.Div().Class("mp-link-meta").
					FlexGrow(ui.Number(1)).
					MinWidth(ui.Px(0)).
					Children(
						ui.Span().Class("mp-engage-label").Text(item.label).
							FontSize(ui.Px(13)).
							FontWeight(ui.Number(600)),
						ui.Span().Class("mp-muted").Text(item.format+" · "+item.cadence),
					),
				ui.Div().Class("mp-engage-stats").Children(download),
			))
	}
	return ui.Div().Class("mp-card mp-stack").Children(
		ui.Div().Class("mp-pagehead").Children(
			ui.Div().Class("mp-pagehead-text").Children(
				chadcn.H1("Reports"),
				chadcn.Muted("Saved views and scheduled exports."),
			),
		),
		chadcn.Card(
			chadcn.CardHeader(
				chadcn.CardTitle("Saved reports"),
				chadcn.CardDescription("Pick a row to preview, download to export."),
			),
			chadcn.CardContent(rows...),
		),
	)
}
