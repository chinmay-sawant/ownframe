package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// kpiPanel renders the five headline inventory KPIs: value, turns, fill
// rate, stockouts, and days of supply. It is a pure function of literals;
// the receiver only matches the panel convention used by the other panels.
func (a *app) kpiPanel() *ui.Element {
	return ui.Div().Class("inv-card inv-kpis").Children(
		inventoryKPICell(
			"Inventory value",
			ui.Text("$2.41M").Class("inv-kpi-value"),
			nil,
			ui.Div().Class("inv-kpi-foot").Children(
				chadcn.Badge("↓ 4.2%"),
				ui.Text("vs Sep · carrying cost down").Class("inv-muted"),
			),
		),
		inventoryKPICell(
			"Inventory turns",
			ui.Div().Class("inv-kpi-row").Children(
				ui.Text("6.8×").Class("inv-kpi-value"),
				ui.Text("12-mo").Class("inv-muted"),
			),
			ui.Div().Class("inv-track").Children(
				ui.Div().Class("inv-track-fill").Width(ui.Percent(68)),
				ui.Div().Class("inv-marker").Left(ui.Percent(68)),
			),
			ui.Div().Class("inv-kpi-foot").Children(
				ui.Text("Target 6.0× · industry median 5.2×").Class("inv-muted"),
			),
		),
		inventoryKPICell(
			"Order fill rate",
			ui.Text("97.2%").Class("inv-kpi-value"),
			nil,
			ui.Div().Class("inv-kpi-foot").Children(
				chadcn.Badge("↑ 0.6 pts"),
				ui.Text("3,918 of 4,031 lines shipped").Class("inv-muted"),
			),
		),
		inventoryKPICell(
			"Stockouts",
			ui.Div().Class("inv-kpi-row").Children(
				ui.Text("7").Class("inv-kpi-value inv-kpi-value--danger"),
				ui.Text("SKUs at zero").Class("inv-muted"),
			),
			nil,
			ui.Div().Class("inv-kpi-foot").Children(
				chadcn.Badge("$18,640", chadcn.BadgeOptions{Variant: chadcn.BadgeDanger}),
				ui.Text("open orders on hold").Class("inv-muted"),
			),
		),
		inventoryKPICell(
			"Avg days of supply",
			ui.Div().Class("inv-kpi-row").Children(
				ui.Text("38").Class("inv-kpi-value"),
				ui.Text("days").Class("inv-muted"),
			),
			ui.Div().Class("inv-segs").Children(
				ui.Div().Class("inv-seg inv-seg--red").Width(ui.Percent(12)),
				ui.Div().Class("inv-seg inv-seg--amber").Width(ui.Percent(18)),
				ui.Div().Class("inv-seg inv-seg--green").Width(ui.Percent(52)),
				ui.Div().Class("inv-seg inv-seg--gray").Width(ui.Percent(18)),
			),
			ui.Div().Class("inv-kpi-foot").Children(
				ui.Text("Under 7 · 7–14 · healthy · overstock").Class("inv-muted"),
			),
		),
	)
}

// inventoryKPICell builds one KPI cell: label, value, an optional middle
// visual (track or segmented bar), and a foot row.
func inventoryKPICell(label string, value, body, foot *ui.Element) *ui.Element {
	cell := ui.Div().Class("inv-kpi").Children(
		ui.Text(label).Class("inv-kpi-label"),
		value,
	)
	if body != nil {
		cell.Children(body)
	}
	if foot != nil {
		cell.Children(foot)
	}
	return cell
}
