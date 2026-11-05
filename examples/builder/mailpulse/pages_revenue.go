//nolint:mnd // Sample revenue fixtures are literal display content: product rows and amounts are fixed demo values.
package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/examples/builder/mailpulse/components"
	"github.com/chinmay-sawant/ownframe/ui"
)

// revenueDotTotal sizes the revenue-by-type dot bars.
const revenueDotTotal = 10

type revenueTypeRow struct {
	id     string
	label  string
	amount string
	filled int
}

func revenueTypeRows() []revenueTypeRow {
	return []revenueTypeRow{
		{id: "paid-deep-dives", label: "Paid deep dives", amount: "$4,204", filled: 10},
		{id: "tool-roundups", label: "Tool roundups", amount: "$2,880", filled: 7},
		{id: "interviews", label: "Interviews", amount: "$2,104", filled: 5},
		{id: "essays", label: "Essays", amount: "$1,480", filled: 4},
		{id: "sponsor-posts", label: "Sponsor posts", amount: "$1,812", filled: 4},
	}
}

type revenueUpgrade struct {
	id     string
	email  string
	plan   string
	amount string
	issue  string
}

func revenueUpgrades() []revenueUpgrade {
	return []revenueUpgrade{
		{id: "maya-studio", email: "maya@studio.co", plan: "Annual", amount: "$129", issue: "Weekly Creator Brief #52"},
		{id: "leo-indie", email: "leo@indiehacker.io", plan: "Monthly", amount: "$12", issue: "AI Tools Worth Trying"},
		{id: "priya-club", email: "priya@newsletter.club", plan: "Annual", amount: "$99", issue: "The Creator Economy Report"},
		{id: "sam-makerlab", email: "sam@makerlab.dev", plan: "Monthly", amount: "$12", issue: "Weekly Creator Brief #51"},
		{id: "ana-creators", email: "ana@creatorshq.com", plan: "Team", amount: "$249", issue: "Sponsor Special #8"},
	}
}

// revenuePage renders the Revenue tab: paid conversions attributed to sends.
// Every data row is selectable through a select:revenue:id:label action.
func (a *app) revenuePage() *ui.Element {
	_ = a
	head := ui.Div().Class("mp-pagehead").Children(
		ui.Div().Class("mp-pagehead-text").Children(
			chadcn.H1("Revenue"),
			chadcn.Muted("Paid conversions attributed to sends."),
		),
	)
	kpis := ui.Div().Class("mp-kpis").Children(
		components.KPICard("MRR", "$12,480", "+8.2%", "Recurring revenue", "green", true),
		components.KPICard("Paid subs", "842", "+3.1%", "Active paying readers", "blue", true),
		components.KPICard("ARPU", "$14.82", "+1.2%", "Per paying subscriber", "purple", true),
		components.KPICard("Churn", "1.8%", "-0.3%", "Cancelled in 30 days", "orange", false),
	)
	typeRows := revenueTypeRows()
	typeItems := make([]*ui.Element, 0, len(typeRows))
	for _, row := range typeRows {
		typeItems = append(typeItems, ui.Div().Class("mp-funnel-row "+selectableRow(false)).
			Attr("data-action", "select:revenue:"+row.id+":"+row.label).
			Children(
				ui.Div().Class("mp-funnel-top").Children(
					ui.Span().Class("mp-funnel-label").Text(row.label),
					ui.Span().Class("mp-funnel-stats").Text(row.amount),
				),
				components.DotBar(row.filled, revenueDotTotal),
			))
	}
	byType := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Revenue by issue type"),
			chadcn.CardDescription("Attributed revenue by issue type."),
		),
		chadcn.CardContent(typeItems...),
	)
	upgradeRows := revenueUpgrades()
	upgradeItems := make([]*ui.Element, 0, len(upgradeRows))
	for _, row := range upgradeRows {
		upgradeItems = append(upgradeItems, ui.Div().Class("mp-linkrow "+selectableRow(false)).
			Attr("data-action", "select:revenue:"+row.id+":"+row.email).
			Children(
				ui.Span().Class("mp-linkicon").Text("$"),
				ui.Div().Class("mp-linkmain").Children(
					ui.Text(row.email).FontWeight(ui.FontWeightBold).FontSize(ui.Px(13)),
					ui.Text(row.issue).FontSize(ui.Px(12)).Color(ui.Hex("#64748b")),
				),
				ui.Div().Class("mp-linkstats").Children(
					ui.Text(row.amount).FontWeight(ui.FontWeightBold),
					ui.Text(row.plan).Color(ui.Hex("#64748b")),
				),
			))
	}
	upgrades := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Recent upgrades"),
			chadcn.CardDescription("Latest paid conversions."),
		),
		chadcn.CardContent(upgradeItems...),
	)
	payout := chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Payout note"),
			chadcn.CardDescription("March payout summary."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-insight").Text("Sponsor posts drive 38% of upgrades despite 9% clicks."),
		),
	)
	return ui.Div().Class("mp-card mp-stack").Children(
		head,
		kpis,
		ui.Div().Class("mp-grid-mid").Children(byType, upgrades),
		payout,
	)
}
