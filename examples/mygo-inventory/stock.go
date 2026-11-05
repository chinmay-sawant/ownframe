// Stock health table: filter pills plus one row per product with the
// stock-vs-reorder bar, days-left coloring, and status.
package main

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"
)

// statusColor colors a status name the way the builder pills do.
func statusColor(status string) string {
	switch status {
	case "Stockout":
		return "#b91c1c"
	case "Reorder":
		return "#b45309"
	case "Overstock":
		return "#475569"
	default:
		return "#15803d"
	}
}

// stockPanel renders the stock health card. The pills and the category
// select both filter the rows; the builder keeps them literal.
func (s *heatApp) stockPanel(c *ui.Context) {
	t := c.Theme()
	s.card(c, func() {
		ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(2).Grow(1).Children(func() {
				ui.Text(c, "Stock health").FontSize(14).Bold()
				ui.Text(c, "Top 8 movers · sorted by days left").FontSize(11).TextColor(t.TextMuted)
			})
			if ui.Button(c, "Export").Clicked() {
				s.picked = "export"
			}
		})
		ui.Row(c).Gap(6).Children(func() {
			for _, pill := range []string{"All SKUs 4,200", "Stockout 7", "Reorder 16", "Overstock 31"} {
				pill := pill
				want := strings.Split(pill, " ")[0]
				if want == "All" {
					want = ""
				}
				b := ui.Button(c, pill)
				if want == s.statusFilter {
					b = b.Background(ui.Hex("#1d5c3f")).TextColor(ui.Hex("#fafaf7"))
				}
				if b.Clicked() {
					s.statusFilter = want
				}
			}
		})
		grid := ui.Grid(c).ColumnTracks(
			ui.Fr(2.2), ui.Fr(1.2), ui.Fixed(64), ui.Fr(1.4), ui.Fixed(56), ui.Fixed(64), ui.Fr(0.9),
		).Gap(6)
		grid.Children(func() {
			for _, h := range []string{"Product", "Category", "On hand", "Stock vs reorder", "Days", "Inbound", "Status"} {
				ui.Text(c, h).FontSize(10).Bold().TextColor(t.TextMuted)
			}
			for _, p := range products {
				if s.statusFilter != "" && p.status != s.statusFilter {
					continue
				}
				if s.category != "All categories" && p.category != s.category {
					continue
				}
				s.stockRow(c, p)
			}
		})
	})
}

// stockRow is one product row: name and code, category, on hand, the bar
// with the reorder marker, colored days left, inbound, and status.
func (s *heatApp) stockRow(c *ui.Context, p product) {
	t := c.Theme()
	ui.Column(c).Gap(1).Children(func() {
		ui.Text(c, p.name).FontSize(12).Bold()
		ui.Text(c, p.code).FontSize(10).TextColor(t.TextMuted)
	})
	ui.Text(c, p.category).FontSize(12)
	ui.Text(c, fmt.Sprint(p.onHand)).FontSize(12)
	ui.Box(c).Height(6).Background(ui.Hex("#eeeeea")).Radius(3).Children(func() {
		ui.Box(c).Absolute().Left(0).Top(0).Bottom(0).
			WidthPercent(float32(p.stock * 100)).Background(ui.Hex("#1d5c3f")).Radius(3)
		ui.Box(c).Absolute().Top(-2).Bottom(-2).
			LeftPercent(float32(p.reorder * 100)).Width(2).Background(ui.Hex("#1c1c1a"))
	})
	days := ui.Text(c, fmt.Sprint(p.daysLeft)).FontSize(12)
	switch {
	case p.daysLeft == 0:
		days = days.TextColor(ui.Hex("#dc2626")).Bold()
	case p.daysLeft <= 9:
		days = days.TextColor(ui.Hex("#d97706"))
	}
	inbound := "—"
	if p.inbound >= 0 {
		inbound = fmt.Sprint(p.inbound)
	}
	ui.Text(c, inbound).FontSize(12)
	ui.Text(c, p.status).FontSize(12).Bold().TextColor(ui.Hex(statusColor(p.status)))
}
