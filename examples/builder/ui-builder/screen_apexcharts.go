package main

import (
	"github.com/chinmay-sawant/ownframe/components/apexcharts"
	"github.com/chinmay-sawant/ownframe/ui"
)

func apexSeries(title string) apexcharts.Options {
	return apexcharts.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []apexcharts.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func apexParts(title string) apexcharts.PartOptions {
	return apexcharts.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []apexcharts.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func apexTree(title string) apexcharts.TreeOptions {
	return apexcharts.TreeOptions{
		Title:       title,
		Description: "Product, services, and support.",
		Nodes: []apexcharts.Node{
			{
				Name: "Product", Value: chartNodeProduct, Color: chartColorBlue,
				Children: []apexcharts.Node{
					{Name: "Core", Value: chartNodeCore, Color: chartColorBlue},
					{Name: "Add-ons", Value: chartNodeAddOn, Color: chartColorViolet},
				},
			},
			{Name: "Services", Value: chartNodeService, Color: chartColorTeal},
			{Name: "Support", Value: chartNodeSupport, Color: chartColorAmber},
		},
	}
}

func apexRange(title string) apexcharts.RangeOptions {
	return apexcharts.RangeOptions{
		Title:       title,
		Description: "Morning through afternoon.",
		Points: []apexcharts.RangePoint{
			{Label: "Mon", Start: chartStartMon, End: chartEndMon},
			{Label: "Tue", Start: chartStartTue, End: chartEndTue},
			{Label: "Wed", Start: chartStartWed, End: chartEndWed},
		},
	}
}

func apexCandles(title string) apexcharts.CandleOptions {
	return apexcharts.CandleOptions{
		Title:       title,
		Description: "Three sessions.",
		Points: []apexcharts.OHLC{
			{Label: "Mon", Open: chartOpenMon, High: chartHighMon, Low: chartLowMon, Close: chartCloseMon},
			{Label: "Tue", Open: chartOpenTue, High: chartHighTue, Low: chartLowTue, Close: chartCloseTue},
			{Label: "Wed", Open: chartOpenWed, High: chartHighWed, Low: chartLowWed, Close: chartCloseWed},
		},
	}
}

func (a *builderApp) apexchartsPanel() *ui.Element {
	return a.chartGallery("ApexCharts", "Line, bar, range, pie, and hierarchical charts.",
		widgetShell("Line", chartSeriesBlurb, apexcharts.Line(apexSeries("Line"))),
		widgetShell("Area", chartSeriesBlurb, apexcharts.Area(apexSeries("Area"))),
		widgetShell("Bar", chartSeriesBlurb, apexcharts.Bar(apexSeries("Bar"))),
		widgetShell("Column", chartSeriesBlurb, apexcharts.Column(apexSeries("Column"))),
		widgetShell("Scatter", chartSeriesBlurb, apexcharts.Scatter(apexSeries("Scatter"))),
		widgetShell("Bubble", chartSeriesBlurb, apexcharts.Bubble(apexSeries("Bubble"))),
		widgetShell("Heatmap", chartSeriesBlurb, apexcharts.Heatmap(apexSeries("Heatmap"))),
		widgetShell("Candlestick", chartCandleBlurb, apexcharts.Candlestick(apexCandles("Candlestick"))),
		widgetShell("Box Plot", chartSeriesBlurb, apexcharts.BoxPlot(apexSeries("Box Plot"))),
		widgetShell("Histogram", chartSeriesBlurb, apexcharts.Histogram(apexSeries("Histogram"))),
		widgetShell("Pie", chartSliceBlurb, apexcharts.Pie(apexParts("Pie"))),
		widgetShell("Donut", chartSliceBlurb, apexcharts.Donut(apexParts("Donut"))),
		widgetShell("Radial Bar", chartSliceBlurb, apexcharts.RadialBar(apexParts("Radial Bar"))),
		widgetShell("Gauge", chartValueBlurb, apexcharts.Gauge("Load", chartGauge)),
		widgetShell("Radar", chartSeriesBlurb, apexcharts.Radar(apexSeries("Radar"))),
		widgetShell("Violin", chartSeriesBlurb, apexcharts.Violin(apexSeries("Violin"))),
		widgetShell("Raincloud", chartSeriesBlurb, apexcharts.Raincloud(apexSeries("Raincloud"))),
		widgetShell("Range Bar", chartRangeBlurb, apexcharts.RangeBar(apexRange("Range Bar"))),
		widgetShell("Range Area", chartRangeBlurb, apexcharts.RangeArea(apexRange("Range Area"))),
		widgetShell("Streamgraph", chartSeriesBlurb, apexcharts.Streamgraph(apexSeries("Streamgraph"))),
		widgetShell("Waterfall", chartSeriesBlurb, apexcharts.Waterfall(apexSeries("Waterfall"))),
		widgetShell("Dumbbell", chartSeriesBlurb, apexcharts.Dumbbell(apexSeries("Dumbbell"))),
		widgetShell("Treemap", chartTreeBlurb, apexcharts.Treemap(apexTree("Treemap"))),
		widgetShell("Polar Area", chartSliceBlurb, apexcharts.PolarArea(apexParts("Polar Area"))),
		widgetShell("Funnel", chartSliceBlurb, apexcharts.Funnel(apexParts("Funnel"))),
		widgetShell("Pyramid", chartSliceBlurb, apexcharts.Pyramid(apexParts("Pyramid"))),
		widgetShell("Sunburst", chartTreeBlurb, apexcharts.Sunburst(apexTree("Sunburst"))),
		widgetShell("Icicle", chartTreeBlurb, apexcharts.Icicle(apexTree("Icicle"))),
		widgetShell("Unit", chartSliceBlurb, apexcharts.Unit(apexParts("Unit"))),
		widgetShell("Waffle", chartSliceBlurb, apexcharts.Waffle(apexParts("Waffle"))),
	)
}
