package main

import (
	"fmt"
	"strconv"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// stockLowDaysThreshold marks a SKU as running low (amber days-left)
// when it has this many or fewer days of cover left.
const stockLowDaysThreshold = 9

func (a *app) stockPanel() *ui.Element {
	rows := make([]*ui.Element, 0, len(sampleSKUs))
	for _, sku := range sampleSKUs {
		if a.statusFilter != "" && sku.Status != a.statusFilter {
			continue
		}
		rows = append(rows, stockRow(sku))
	}
	return ui.Div().Class("inv-card").Children(
		ui.Div().Class("inv-card-head").Children(
			ui.Div().Children(
				ui.Text("Stock health").FontWeight(ui.FontWeightBold).FontSize(ui.Px(14)),
				ui.Text("Top 8 movers · sorted by days left").Class("app-muted"),
			),
			chadcn.Button("Export", chadcn.ButtonOptions{Variant: chadcn.ButtonOutlined, Size: chadcn.ButtonSmall}).
				Attr("data-action", "export:run"),
		),
		a.statusPills(),
		ui.Div().Class("inv-table-wrap").Children(
			chadcn.Table(
				chadcn.TableHeader(chadcn.TableRow(
					chadcn.TableHead("PRODUCT"), chadcn.TableHead("CATEGORY"),
					chadcn.TableHead("ON HAND"), chadcn.TableHead("STOCK VS REORDER POINT"),
					chadcn.TableHead("DAYS LEFT ↑"), chadcn.TableHead("INBOUND"),
					chadcn.TableHead("STATUS"),
				)),
				chadcn.TableBody(rows...),
			),
		),
	)
}

// statusPills is the mygo-style filter row: All SKUs plus one pill per
// non-healthy status. Clicking a pill sets a.statusFilter and the table
// shows only that status; the active pill carries the dark-green fill,
// mirroring the mygo reference highlight.
func (a *app) statusPills() *ui.Element {
	pills := []struct{ label, filter, id string }{
		{"All SKUs 4,200", "", "all"},
		{"Stockout 7", "Stockout", "stockout"},
		{"Reorder 16", "Reorder", "reorder"},
		{"Overstock 31", "Overstock", "overstock"},
	}
	items := make([]*ui.Element, 0, len(pills))
	for _, pill := range pills {
		el := ui.Span().ID("inv-pill-"+pill.id).Class("inv-pill").
			Attr("data-action", "status:"+pill.filter).
			Text(pill.label)
		if a.statusFilter == pill.filter {
			el.BackgroundColor(ui.Hex("#1d5c3f")).Color(ui.Hex("#fafaf7")).FontWeight(ui.FontWeightBold)
		}
		items = append(items, el)
	}
	return ui.Div().Class("inv-filters").Children(items...)
}

func stockRow(sku sku) *ui.Element {
	inbound := "—"
	if sku.Inbound >= 0 {
		inbound = strconv.Itoa(sku.Inbound)
	}
	days := ui.Text(fmt.Sprintf("%d", sku.DaysLeft))
	switch {
	case sku.DaysLeft == 0:
		days.Color(ui.Hex("#dc2626")).FontWeight(ui.FontWeightBold)
	case sku.DaysLeft <= stockLowDaysThreshold:
		days.Color(ui.Hex("#d97706"))
	}
	return chadcn.TableRow(
		chadcn.TableCell("", ui.Div().Children(
			ui.Text(sku.Name).FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
			ui.Text(sku.Code).Class("app-muted"),
		)),
		chadcn.TableCell(sku.Category),
		chadcn.TableCell(strconv.Itoa(sku.OnHand)),
		chadcn.TableCell("", ui.Div().Class("inv-bar").Children(
			ui.Div().Class("inv-bar-fill "+stockFillClass(sku.Status)).Width(ui.Percent(sku.StockPct*100)),
			ui.Div().Class("inv-bar-marker").Left(ui.Percent(sku.ReorderPct*100)),
		)),
		chadcn.TableCell("", days),
		chadcn.TableCell(inbound),
		chadcn.TableCell("", ui.Span().Class(stockPillClass(sku.Status)).Text(sku.Status)),
	).ID("stock-row-" + sku.Code)
}

func stockPillClass(status string) string {
	switch status {
	case "Stockout":
		return "pill-stockout"
	case "Reorder":
		return "pill-reorder"
	case "Overstock":
		return "pill-overstock"
	default:
		return "pill-healthy"
	}
}

func stockFillClass(status string) string {
	switch status {
	case "Stockout":
		return "inv-bar-fill--stockout"
	case "Reorder":
		return "inv-bar-fill--reorder"
	case "Overstock":
		return "inv-bar-fill--overstock"
	default:
		return "inv-bar-fill--healthy"
	}
}
