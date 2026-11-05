package main

import (
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/examples/builder/mailpulse/components"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Heat intensity levels behind the Best time to send grid, from quiet (0)
// to peak (4).
const (
	heatLevelQuiet    = 0
	heatLevelLow      = 1
	heatLevelModerate = 2
	heatLevelHigh     = 3
	heatLevelPeak     = 4
)

// heatPinnedDay and heatPinnedCol mark the pinned best send slot (Tue 9a)
// the readout falls back to until the reader picks a cell.
const (
	heatPinnedDay = 1
	heatPinnedCol = 3
)

// heatCol* name the labeled hour columns of the 18-column grid.
const (
	heatCol6a  = 0
	heatCol8a  = 2
	heatCol10a = 4
	heatCol12p = 6
	heatCol2p  = 8
	heatCol4p  = 10
	heatCol6p  = 12
	heatCol8p  = 14
	heatCol10p = 16
)

// heatCellID names one heatmap cell so hover and click handlers can map it
// back to a weekday and hour column.
func heatCellID(day, col int) string {
	return "mp-heat-" + strconv.Itoa(day) + "-" + strconv.Itoa(col)
}

// parseHeatCellID recovers the weekday and column from a heat cell id.
func parseHeatCellID(id string) (int, int, bool) {
	rest, ok := strings.CutPrefix(id, "mp-heat-")
	if !ok {
		return 0, 0, false
	}
	day, col, ok := strings.Cut(rest, "-")
	if !ok {
		return 0, 0, false
	}
	d, err := strconv.Atoi(day)
	if err != nil {
		return 0, 0, false
	}
	c, err := strconv.Atoi(col)
	if err != nil {
		return 0, 0, false
	}
	if d < 0 || d > 6 || c < 0 || c > 17 {
		return 0, 0, false
	}
	return d, c, true
}

// heatLevelRate maps an intensity level to a representative open rate.
func heatLevelRate(level int) string {
	switch level {
	case heatLevelPeak:
		return "68%"
	case heatLevelHigh:
		return "54%"
	case heatLevelModerate:
		return "38%"
	case heatLevelLow:
		return "22%"
	default:
		return "8%"
	}
}

// heatLevelWord names an intensity level for the readout line.
func heatLevelWord(level int) string {
	switch level {
	case heatLevelPeak:
		return "peak"
	case heatLevelHigh:
		return "high"
	case heatLevelModerate:
		return "moderate"
	case heatLevelLow:
		return "low"
	default:
		return "quiet"
	}
}

// heatCellLabel builds the readout line for one cell, e.g.
// "Tue 9a · high open rate · about 54%".
func heatCellLabel(day, col int) string {
	days := heatDays()
	hours := heatHours()
	if day < 0 || day >= len(days) || col < 0 || col >= len(hours) {
		return "Hover a cell to see the open rate."
	}
	level := heatValue(day, col)
	return days[day] + " " + hours[col] + " · " + heatLevelWord(level) +
		" open rate · about " + heatLevelRate(level)
}

func (a *app) heatReadout() string {
	if a.heatDay >= 0 && a.heatCol >= 0 {
		return heatCellLabel(a.heatDay, a.heatCol)
	}
	// Default to the pinned best slot until the reader picks a cell.
	return heatCellLabel(heatPinnedDay, heatPinnedCol)
}

// heatValueFor returns the heatmap intensity for a range window. The 30-
// and 90-day windows share the full sample; the 7-day window dampens
// non-peak cells by one step so the smaller sample reads cooler while the
// pinned Tue/Thu 9a peaks stay lit.
func heatValueFor(day, col, days int) int {
	base := heatValue(day, col)
	if days == 7 && base > 0 && base < 4 {
		return base - 1
	}
	return base
}

func (a *app) heatmapCard() *ui.Element {
	const cols = 18
	days := heatDays()
	hourAt := func(col int) string {
		switch col {
		case heatCol6a:
			return "6a"
		case heatCol8a:
			return "8a"
		case heatCol10a:
			return "10a"
		case heatCol12p:
			return "12p"
		case heatCol2p:
			return "2p"
		case heatCol4p:
			return "4p"
		case heatCol6p:
			return "6p"
		case heatCol8p:
			return "8p"
		case heatCol10p:
			return "10p"
		default:
			return ""
		}
	}
	gridCells := make([]*ui.Element, 0, (len(days)+1)*(cols+1))
	for day, label := range days {
		gridCells = append(gridCells, ui.Span().Class("mp-heat-day").Text(label))
		for col := 0; col < cols; col++ {
			gridCells = append(gridCells, components.HeatCellID(heatCellID(day, col), heatValueFor(day, col, a.rangeDays())))
		}
	}
	gridCells = append(gridCells, ui.Span().Class("mp-heat-day").Text(""))
	for col := 0; col < cols; col++ {
		gridCells = append(gridCells, ui.Span().Class("mp-heat-hour").Text(hourAt(col)))
	}
	return chadcn.Card(
		chadcn.CardHeader(
			chadcn.CardTitle("Best time to send"),
			chadcn.CardDescription("Open rate by weekday and hour, "+a.rangeWindow()+". Hover or tap a cell."),
		),
		chadcn.CardContent(
			ui.Div().Class("mp-heat-badges").Children(
				chadcn.Badge("✦ Best: Tue 9 AM"),
				chadcn.Badge("Open rate", chadcn.BadgeOptions{Variant: chadcn.BadgeOutline}),
			),
			ui.Div().Class("mp-heat-grid").Children(gridCells...),
			ui.Div().ID("mp-heat-readout").Class("mp-heat-readout").Text(a.heatReadout()),
			ui.Div().Class("mp-heat-legend").Children(
				ui.Span().Class("mp-heat-legend-label").Text("Lower"),
				ui.Div().Class("mp-heat-swatches").Children(
					components.HeatCell(heatLevelQuiet),
					components.HeatCell(heatLevelLow),
					components.HeatCell(heatLevelModerate),
					components.HeatCell(heatLevelHigh),
					components.HeatCell(heatLevelPeak),
				),
				ui.Span().Class("mp-heat-legend-label").Text("Higher open rate"),
			),
		),
	)
}
