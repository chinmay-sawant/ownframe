// Inbound purchase orders and inventory value by category.
package main

import (
	"github.com/egoist/mygo/ui"
)

// inboundPanel renders the inbound purchase-order card: open-order
// summary plus one timeline row per order.
func (s *heatApp) inboundPanel(c *ui.Context) {
	t := c.Theme()
	s.card(c, func() {
		ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(2).Grow(1).Children(func() {
				ui.Text(c, "Inbound purchase orders").FontSize(14).Bold()
				ui.Text(c, "4 open · $186,420").FontSize(11).TextColor(t.TextMuted)
			})
			ui.Text(c, "View all").FontSize(12).TextColor(t.TextMuted)
		})
		ui.Column(c).Gap(8).Children(func() {
			for index, po := range purchaseOrders {
				s.inboundRow(c, index == 0, po)
			}
		})
	})
}

// inboundRow is one purchase order: a status dot, vendor and PO number,
// the detail line, and the amount with its badge.
func (s *heatApp) inboundRow(c *ui.Context, first bool, po inboundPO) {
	t := c.Theme()
	dot := ui.Hex("#d6d3cb")
	if first {
		dot = ui.Hex("#1d5c3f")
	}
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Box(c).Size(8, 8).Radius(4).Background(dot)
		ui.Column(c).Gap(2).Grow(1).Children(func() {
			ui.Row(c).Gap(6).Children(func() {
				ui.Text(c, po.vendor).FontSize(12).Bold()
				ui.Text(c, po.poNo).FontSize(11).TextColor(t.TextMuted)
			})
			ui.Text(c, po.detail+" · "+po.eta).FontSize(11).TextColor(t.TextMuted)
		})
		ui.Column(c).Gap(2).Children(func() {
			ui.Text(c, po.amount).FontSize(12).Bold()
			ui.Text(c, po.badge).FontSize(10).TextColor(ui.Hex(inboundBadgeColor(po.badge)))
		})
	})
}

// inboundBadgeColor tints the badge the way the builder pills do.
func inboundBadgeColor(badge string) string {
	switch badge {
	case "Arriving today":
		return "#1d5c3f"
	case "On the water":
		return "#1d4ed8"
	default:
		return "#5b6b5f"
	}
}

// categoryPanel shows inventory value split across categories as a
// stacked bar with a legend, from the category shares.
func (s *heatApp) categoryPanel(c *ui.Context) {
	t := c.Theme()
	s.card(c, func() {
		ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(2).Grow(1).Children(func() {
				ui.Text(c, "Inventory value by category").FontSize(14).Bold()
				ui.Text(c, "at landed cost").FontSize(11).TextColor(t.TextMuted)
			})
			ui.Text(c, "$2,410,000").FontSize(18).Bold()
		})
		ui.Row(c).Gap(0).Children(func() {
			for i, cat := range categories {
				ui.Box(c).Grow(float32(cat.share)).Height(12).
					Background(ui.Hex(catColors[i%len(catColors)]))
			}
		})
		ui.Column(c).Gap(4).Children(func() {
			for i, cat := range categories {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(10, 10).Radius(3).Background(ui.Hex(catColors[i%len(catColors)]))
					ui.Text(c, cat.label).FontSize(12).Grow(1)
					ui.Text(c, cat.amount).FontSize(12).Bold()
					ui.Text(c, cat.pct).FontSize(11).TextColor(t.TextMuted)
				})
			}
		})
	})
}
