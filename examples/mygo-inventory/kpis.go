// KPI strip: the five headline inventory indicators.
package main

import (
	"github.com/egoist/mygo/ui"
)

// kpiPanel renders the five headline KPIs: value, turns, fill rate,
// stockouts, and days of supply. Literals mirror builder/inventory kpis.go.
func (s *heatApp) kpiPanel(c *ui.Context) {
	t := c.Theme()
	s.card(c, func() {
		ui.Row(c).Gap(0).Children(func() {
			s.kpiCell(c, "Inventory value", func() {
				ui.Text(c, "$2.41M").FontSize(22).Bold()
				ui.Row(c).Gap(6).Children(func() {
					ui.Badge(c, "↓ 4.2%")
					ui.Text(c, "vs Sep · carrying cost down").FontSize(10).TextColor(t.TextMuted)
				})
			})
			s.kpiCell(c, "Inventory turns", func() {
				ui.Row(c).Gap(6).Children(func() {
					ui.Text(c, "6.8×").FontSize(22).Bold()
					ui.Text(c, "12-mo").FontSize(10).TextColor(t.TextMuted)
				})
				ui.Box(c).Height(6).Background(ui.Hex("#eeeeea")).Radius(3).Children(func() {
					ui.Box(c).Absolute().Left(0).Top(0).Bottom(0).
						WidthPercent(68).Background(ui.Hex("#1d5c3f")).Radius(3)
				})
				ui.Text(c, "Target 6.0× · industry median 5.2×").FontSize(10).TextColor(t.TextMuted)
			})
			s.kpiCell(c, "Order fill rate", func() {
				ui.Text(c, "97.2%").FontSize(22).Bold()
				ui.Row(c).Gap(6).Children(func() {
					ui.Badge(c, "↑ 0.6 pts")
					ui.Text(c, "3,918 of 4,031 lines shipped").FontSize(10).TextColor(t.TextMuted)
				})
			})
			s.kpiCell(c, "Stockouts", func() {
				ui.Row(c).Gap(6).Children(func() {
					ui.Text(c, "7").FontSize(22).Bold().TextColor(ui.Hex("#b91c1c"))
					ui.Text(c, "SKUs at zero").FontSize(10).TextColor(t.TextMuted)
				})
				ui.Row(c).Gap(6).Children(func() {
					ui.Badge(c, "$18,640")
					ui.Text(c, "open orders on hold").FontSize(10).TextColor(t.TextMuted)
				})
			})
			s.kpiCell(c, "Avg days of supply", func() {
				ui.Row(c).Gap(6).Children(func() {
					ui.Text(c, "38").FontSize(22).Bold()
					ui.Text(c, "days").FontSize(10).TextColor(t.TextMuted)
				})
				ui.Row(c).Gap(2).Children(func() {
					for _, seg := range []struct {
						share float32
						color string
					}{
						{12, "#dc2626"}, {18, "#f59e0b"}, {52, "#1d5c3f"}, {18, "#d6d3cb"},
					} {
						ui.Box(c).Grow(seg.share).Height(6).Background(ui.Hex(seg.color)).Radius(3)
					}
				})
				ui.Text(c, "Under 7 · 7–14 · healthy · overstock").FontSize(10).TextColor(t.TextMuted)
			})
		})
	})
}

// kpiCell is one KPI cell: an uppercase label plus the value block.
func (s *heatApp) kpiCell(c *ui.Context, label string, build func()) {
	t := c.Theme()
	ui.Column(c).Gap(4).Grow(1).Children(func() {
		ui.Text(c, label).FontSize(10).TextColor(t.TextMuted)
		build()
	})
}

// card is a white rounded panel around build.
func (s *heatApp) card(c *ui.Context, build func()) {
	ui.Box(c).Background(ui.Hex("#ffffff")).Radius(10).Padding(12).Children(build)
}
