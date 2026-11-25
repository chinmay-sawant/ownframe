package main

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

func statNum(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// statusBar shows the mode, the aggregates of the selection, and the zoom.
func (a *app) statusBar() *ui.Element {
	mode := "Ready"
	if a.editing {
		mode = "Edit"
	}
	bar := ui.Div().ID("xl-status").Class("xl-status").Children(
		ui.Span().ID("xl-status-mode").Text(mode),
		ui.Div().Class("xl-status-spacer"),
	)
	if count, numCount, sum, avg, _, _ := a.sheet().Stats(a.sel); numCount > 0 && a.sel.Min != a.sel.Max {
		bar.Children(
			ui.Span().ID("xl-stat-avg").Class("xl-status-stat").Text("Average: "+statNum(round4(avg))),
			ui.Span().ID("xl-stat-count").Class("xl-status-stat").Text("Count: "+itoa(count)),
			ui.Span().ID("xl-stat-sum").Class("xl-status-stat").Text("Sum: "+statNum(round4(sum))),
		)
	} else if count > 0 && a.sel.Min != a.sel.Max {
		bar.Children(ui.Span().ID("xl-stat-count").Class("xl-status-stat").Text("Count: " + itoa(count)))
	}
	bar.Children(ui.Span().ID("xl-status-zoom").Class("xl-status-stat").Text(itoa(a.zoomPct()) + "%"))
	return bar
}

func round4(f float64) float64 {
	const (
		k         = 10000
		roundHalf = 0.5
	)
	if f >= 0 {
		return float64(int64(f*k+roundHalf)) / k
	}
	return float64(int64(f*k-roundHalf)) / k
}
