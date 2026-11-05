// Bin utilization grid: aisles A-J by rack level. Hovering a cell shows
// its data in the note line, the way the builder patches #bin-detail;
// every cell also carries the same text as a tooltip.
package main

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// binsPanel renders the bin utilization card: title, the hover note,
// the 11-column grid, and the legend.
func (s *heatApp) binsPanel(c *ui.Context) {
	t := c.Theme()
	// hover fills in below, as the cells build; the note shows it from the
	// next frame on, while hovering keeps frames coming.
	hover := ""
	s.card(c, func() {
		ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(2).Grow(1).Children(func() {
				ui.Text(c, "Bin utilization").FontSize(14).Bold()
				ui.Text(c, "Aisles A–J by rack level · % of cubic capacity").FontSize(11).TextColor(t.TextMuted)
			})
			ui.Text(c, "77% full").FontSize(13).Bold()
		})
		note := binSummary()
		if hover != "" {
			note = hover
		}
		ui.Text(c, note).FontSize(12)
		tracks := []ui.Track{ui.Fixed(32)}
		for range 10 {
			tracks = append(tracks, ui.Fr(1))
		}
		grid := ui.Grid(c).ColumnTracks(tracks...).Gap(4)
		grid.Children(func() {
			ui.Box(c)
			for aisle := range 10 {
				ui.Box(c).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
					ui.Text(c, fmt.Sprintf("%c", 'A'+aisle)).FontSize(11).TextColor(t.TextMuted)
				})
			}
			for li, level := range binLevels {
				ui.Box(c).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
					ui.Text(c, level).FontSize(11).TextColor(t.TextMuted)
				})
				for aisle := range 10 {
					hover = s.binCell(c, hover, li, aisle, level)
				}
			}
		})
		ui.Row(c).Gap(8).Children(func() {
			for _, stop := range binStops {
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(12, 12).Radius(3).Background(ui.Hex(stop.bg))
					ui.Text(c, stop.label+" "+stop.scale).FontSize(11).TextColor(t.TextMuted)
				})
			}
		})
	})
}

// binCell is one heat cell. It reports hover back by returning the
// hovered title, or the incoming hover when untouched.
func (s *heatApp) binCell(c *ui.Context, hover string, li, aisle int, level string) string {
	value := bins[li][aisle]
	bg, fg := binColors(value)
	title := binCellTitle(aisle, level, value)
	cell := ui.Box(c).Background(ui.Hex(bg)).
		AlignItems(ui.Center).Justify(ui.Center).
		Radius(3).Tooltip(title)
	cell.Children(func() {
		ui.Text(c, fmt.Sprint(value)).FontSize(11).TextColor(ui.Hex(fg))
	})
	if cell.Hovered() {
		return title
	}
	return hover
}
