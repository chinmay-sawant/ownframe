package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// binsPanel renders the bin utilization heat grid: aisles A–J by rack level.
func (a *app) binsPanel() *ui.Element {
	return ui.Div().Class("inv-card").Children(chadcn.Card(
		ui.Div().Class("app-row app-row--between").Children(
			ui.Div().Class("app-col").Gap(ui.Px(3)).Children(
				chadcn.CardTitle("Bin utilization"),
				chadcn.CardDescription("Aisles A–J by rack level · % of cubic capacity"),
			),
			ui.Text("77% full").FontWeight(ui.FontWeightBold),
		),
		ui.Div().ID("bin-detail").Class("inv-bins-note").Text(binSummary()),
		chadcn.CardContent(a.binsGrid()),
		ui.Div().Class("app-row app-row--between").Children(
			a.binsLegend(),
			chadcn.Button("Suggest re-slot · 46 SKUs →", chadcn.ButtonOptions{Variant: chadcn.ButtonText, Size: chadcn.ButtonSmall}),
		),
	).Padding(ui.Px(16)).Gap(ui.Px(10)))
}

// aisleMax is the largest aisle index aisleLetter accepts (aisle Z).
const aisleMax = 25

// aisleLetter returns the A–Z label for an aisle index. Out-of-range
// indexes clamp so the int-to-rune conversion cannot overflow.
func aisleLetter(aisle int) string {
	if aisle < 0 {
		aisle = 0
	}
	if aisle > aisleMax {
		aisle = aisleMax
	}
	return string(rune('A' + aisle))
}

func (a *app) binsGrid() *ui.Element {
	grid := ui.Div().Class("inv-bins-grid").
		GridTemplateColumns(ui.RawValue("32px repeat(10, minmax(0, 1fr))")).
		Gap(ui.Px(4))
	header := []*ui.Element{ui.Div().Class("inv-bin-head")}
	for aisle := 0; aisle < 10; aisle++ {
		header = append(header, ui.Div().Class("inv-bin-head").Text(aisleLetter(aisle)))
	}
	grid.Children(header...)
	for _, row := range sampleBins {
		cells := []*ui.Element{ui.Div().Class("inv-bin-head").Text(row.Level)}
		for aisle, value := range row.Cells {
			cells = append(cells, a.binCell(aisle, row.Level, value))
		}
		grid.Children(cells...)
	}
	return grid
}

func (a *app) binCell(aisle int, level string, value int) *ui.Element {
	bg, fg := binColors(value)
	label := "Aisle " + aisleLetter(aisle) + " · level " + level + " · " + strconv.Itoa(value) + "% full"
	return ui.Div().ID("bin-"+aisleLetter(aisle)+"-"+level).Class("inv-bin").
		BackgroundColor(ui.Hex(bg)).
		Color(ui.Hex(fg)).
		Attr("aria-label", label).
		Text(strconv.Itoa(value))
}

// binSummary is the hottest/coldest aisle line shared by the panel and hover.
func binSummary() string {
	hotName, hotAvg := hottestAisle()
	coldName, coldAvg := coldestAisle()
	return fmt.Sprintf("Hottest aisle %s · %d%% avg · Coldest aisle %s · %d%% avg", hotName, hotAvg, coldName, coldAvg)
}

// binTitleByID parses a bin cell ID ("bin-<A>-<L>", e.g. "bin-C-L3") and
// returns the cell title for the matching sampleBins value.
func binTitleByID(id string) (string, bool) {
	rest, ok := strings.CutPrefix(id, "bin-")
	if !ok {
		return "", false
	}
	parts := strings.Split(rest, "-")
	if len(parts) != 2 || len(parts[0]) != 1 {
		return "", false
	}
	aisle := int(parts[0][0] - 'A')
	if aisle < 0 || aisle > 9 {
		return "", false
	}
	level := parts[1]
	if len(level) != 2 || level[0] != 'L' || level[1] < '1' || level[1] > '5' {
		return "", false
	}
	for _, row := range sampleBins {
		if row.Level == level {
			value := row.Cells[aisle]
			return binCellTitle(aisle, level, value), true
		}
	}
	return "", false
}

func (a *app) binsLegend() *ui.Element {
	legend := ui.Div().Class("app-row").Gap(ui.Px(4))
	for _, stop := range binStops {
		legend.Children(ui.Div().Class("inv-legend-stop").Children(
			ui.Div().Class("inv-bin-key").BackgroundColor(ui.Hex(stop.Bg)),
			ui.Text(stop.Label+" "+stop.Range).Class("app-muted"),
		))
	}
	return legend
}
