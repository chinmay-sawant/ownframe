package main

import (
	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// inboundPanel renders the inbound purchase-order card: a header with the
// open-order summary plus a timeline list of rows from samplePOs. Clicking
// a row selects that order (po:view:<PONo>); View all clears the selection.
func (a *app) inboundPanel() *ui.Element {
	rows := make([]*ui.Element, 0, len(samplePOs))
	for index, po := range samplePOs {
		rows = append(rows, inboundPORow(a.selectedPO == po.PONo, index, po))
	}
	return ui.Div().Class("inv-card").Children(
		chadcn.Card(
			chadcn.CardHeader(
				ui.Div().Class("app-row app-row--between").Children(
					ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
						chadcn.CardTitle("Inbound purchase orders"),
						chadcn.CardDescription(a.inboundSummary()),
					),
					chadcn.Button("View all", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}).
						Attr("data-action", "po:viewall"),
				),
			),
			chadcn.CardContent(
				ui.Div().Class("app-col").Gap(ui.Px(8)).Children(rows...),
			),
		).Padding(ui.Px(16)).Gap(ui.Px(10)),
	)
}

// inboundSummary is the header line: the selected order's detail, or the
// open-order totals when nothing is selected.
func (a *app) inboundSummary() string {
	for _, po := range samplePOs {
		if po.PONo == a.selectedPO {
			return po.PONo + " · " + po.Vendor + " · " + po.Amount
		}
	}
	return "4 open · $186,420"
}

// inboundPORow renders one purchase order as a selectable timeline row: a
// status dot, the vendor and PO number, the detail line, and the amount
// with its badge. The selected row carries the highlighted fill.
func inboundPORow(selected bool, index int, po inboundPO) *ui.Element {
	dot := ui.Span().Class("inv-dot")
	if index == 0 {
		dot = ui.Span().Class("inv-dot inv-dot--filled")
	}
	row := ui.Div().Class("inv-po").Attr("data-action", "po:view:"+po.PONo)
	if selected {
		row.BackgroundColor(ui.Hex("#e8f0e8")).BorderRadius(ui.Px(8))
	}
	return row.Children(
		chadcn.Item(chadcn.ItemOptions{
			Media: dot,
			Content: ui.Div().Class("app-col").Gap(ui.Px(2)).Children(
				ui.Div().Class("app-row").Children(
					ui.Text(po.Vendor).FontWeight(ui.FontWeightBold).FontSize(ui.Px(12)),
					ui.Text(po.PONo).Class("app-muted"),
				),
				ui.Text(po.Detail).Class("app-muted").FontSize(ui.Px(11)),
			),
			Actions: ui.Div().Class("inv-po-side").Children(
				ui.Text(po.Amount).FontWeight(ui.FontWeightBold),
				inboundPOBadge(po.Badge),
			),
		}).Padding(ui.Edges(ui.Px(10), ui.Px(12))),
	)
}

// inboundPOBadge renders the status pill: dark-green filled for arriving
// orders, gray-green tinted for in-transit ones, blue tinted on the water.
func inboundPOBadge(badge string) *ui.Element {
	class := "inv-badge inv-badge--transit"
	switch badge {
	case "Arriving today":
		class = "inv-badge inv-badge--arriving"
	case "On the water":
		class = "inv-badge inv-badge--water"
	}
	return ui.Span().Class(class).Text(badge)
}
