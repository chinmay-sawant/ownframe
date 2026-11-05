package main

import (
	"github.com/chinmay-sawant/ownframe/components/chartjs"
	"github.com/chinmay-sawant/ownframe/ui"
	"github.com/chinmay-sawant/ownframe/ui/theme"
)

// Chart gallery sample data. Every library shows the same week, share, or flow.
const (
	chartColorBlue   = "#4677ed"
	chartColorTeal   = "#59d4b8"
	chartColorAmber  = "#e5a54b"
	chartColorViolet = "#9b7ede"

	chartSeriesBlurb = "Two series across three labels."
	chartSliceBlurb  = "Three slices."
	chartTreeBlurb   = "Three nodes."
	chartFlowBlurb   = "Three nodes and two links."
	chartCandleBlurb = "Three open-high-low-close points."
	chartRangeBlurb  = "Three start and end points."
	chartDayBlurb    = "Three days."
	chartTrackBlurb  = "Three status cells."
	chartValueBlurb  = "One value."

	chartMon  = 24.0
	chartTue  = 18.0
	chartWed  = 31.0
	chartMonB = 12.0
	chartTueB = 16.0
	chartWedB = 22.0

	chartDesktop = 46.0
	chartMobile  = 32.0
	chartTablet  = 22.0

	chartNodeProduct = 40.0
	chartNodeCore    = 22.0
	chartNodeAddOn   = 18.0
	chartNodeService = 28.0
	chartNodeSupport = 16.0
	chartLinkIn      = 25.0
	chartLinkOut     = 15.0

	chartOpenMon  = 20.0
	chartHighMon  = 28.0
	chartLowMon   = 18.0
	chartCloseMon = 26.0
	chartOpenTue  = 26.0
	chartHighTue  = 30.0
	chartLowTue   = 22.0
	chartCloseTue = 24.0
	chartOpenWed  = 24.0
	chartHighWed  = 32.0
	chartLowWed   = 23.0
	chartCloseWed = 31.0

	chartStartMon = 8.0
	chartEndMon   = 12.0
	chartStartTue = 9.0
	chartEndTue   = 15.0
	chartStartWed = 7.0
	chartEndWed   = 11.0

	chartDayMon = 4.0
	chartDayTue = 7.0
	chartDayWed = 5.0

	chartGauge = 72.0
)

func (a *builderApp) chartGallery(title, intro string, cards ...*ui.Element) *ui.Element {
	children := make([]*ui.Element, 0, len(cards)+1)
	children = append(children, ui.Div().Children(
		a.panelHeading(title),
		ui.P().Text(intro).Margin(ui.Edges(ui.Px(7), ui.Px(0), ui.Px(0))).Color(theme.Muted()).FontSize(ui.Px(12)),
	))
	children = append(children, cards...)
	return ui.Section().Flex(ui.Number(1)).MinWidth(ui.Px(0)).MinHeight(ui.Px(0)).
		Display(ui.DisplayFlex).FlexDirection(ui.FlexDirectionColumn).Gap(ui.Px(12)).
		Children(children...)
}

func chartjsSeries(title string) chartjs.Options {
	return chartjs.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []chartjs.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func chartjsParts(title string) chartjs.PartOptions {
	return chartjs.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []chartjs.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func (a *builderApp) chartjsPanel() *ui.Element {
	return a.chartGallery("Chart.js", "Bar, line, pie, doughnut, radar, polar area, bubble, and scatter.",
		widgetShell("Bar", chartSeriesBlurb, chartjs.Bar(chartjsSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, chartjs.Line(chartjsSeries("Line"))),
		widgetShell("Pie", chartSliceBlurb, chartjs.Pie(chartjsParts("Pie"))),
		widgetShell("Doughnut", chartSliceBlurb, chartjs.Doughnut(chartjsParts("Doughnut"))),
		widgetShell("Radar", chartSeriesBlurb, chartjs.Radar(chartjsSeries("Radar"))),
		widgetShell("Polar Area", chartSliceBlurb, chartjs.PolarArea(chartjsParts("Polar Area"))),
		widgetShell("Bubble", chartSeriesBlurb, chartjs.Bubble(chartjsSeries("Bubble"))),
		widgetShell("Scatter", chartSeriesBlurb, chartjs.Scatter(chartjsSeries("Scatter"))),
	)
}
