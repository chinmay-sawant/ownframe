// Product heat map: the 8 separate products against aisles A-J, drawn
// with ui/chart. A cell scales its aisle's average fill by the product's
// own stock health and animates background changes.
package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/chart"
)

// heatPanel renders the stock-by-product heat card with the selection
// readout. Rows run bottom-up in chart space, so the first product maps
// to the top band.
func (s *heatApp) heatPanel(c *ui.Context) {
	t := c.Theme()
	aisles := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J"}
	rows := make([]string, len(products))
	for i, p := range products {
		rows[len(products)-1-i] = p.short
	}
	x := chart.NewBand(aisles...).Padding(0.08, 0.04)
	y := chart.NewBand(rows...).Padding(0.08, 0.04)
	s.card(c, func() {
		ui.Row(c).Gap(12).AlignItems(ui.End).Children(func() {
			ui.Column(c).Gap(2).Grow(1).Children(func() {
				ui.Text(c, "Stock heat by product").FontSize(14).Bold()
				ui.Text(c, "8 products × 10 aisles · cell = aisle profile × product stock health").FontSize(11).TextColor(t.TextMuted)
			})
		})
		ui.Text(c, s.selectedText()).FontSize(12)
		chart.Plot(c, chart.Axes{Left: y.Ticks(), Bottom: x.Ticks()}, func(area ui.Element) {
			for pi := range products {
				for ai := range aisles {
					v := cellValue(pi, ai)
					bg, fg := binColors(v)
					cell := chart.Rect(c, x.Band(ai), y.Band(len(products)-1-pi)).
						Background(ui.Hex(bg)).
						AlignItems(ui.Center).
						Justify(ui.Center).
						Label(cellLabel(pi, ai)).
						Tooltip(cellTip(pi, ai)).
						Transition(ui.ElementTransition{Duration: 250 * time.Millisecond, Colors: true})
					cell.Children(func() {
						ui.Text(c, fmt.Sprint(v)).FontSize(11).TextColor(ui.Hex(fg))
					})
					if cell.Clicked() {
						s.selP, s.selA = pi, ai
					}
				}
			}
		}).Height(360).Gap(8)
	})
}

func (s *heatApp) selectedText() string {
	if s.selP < 0 || s.selA < 0 {
		return "Selected: none - click a cell"
	}
	p := products[s.selP]
	v := cellValue(s.selP, s.selA)
	return fmt.Sprintf("Selected: %s · Aisle %c · %d%% · %s",
		p.name, 'A'+s.selA, v, binStatus(v))
}
