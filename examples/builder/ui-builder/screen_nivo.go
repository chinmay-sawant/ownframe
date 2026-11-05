package main

import (
	"github.com/chinmay-sawant/ownframe/components/nivo"
	"github.com/chinmay-sawant/ownframe/ui"
)

func nivoSeries(title string) nivo.Options {
	return nivo.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []nivo.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func nivoParts(title string) nivo.PartOptions {
	return nivo.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []nivo.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func nivoTree(title string) nivo.TreeOptions {
	return nivo.TreeOptions{
		Title:       title,
		Description: "Product, services, and support.",
		Nodes: []nivo.Node{
			{
				Name: "Product", Value: chartNodeProduct, Color: chartColorBlue,
				Children: []nivo.Node{
					{Name: "Core", Value: chartNodeCore, Color: chartColorBlue},
					{Name: "Add-ons", Value: chartNodeAddOn, Color: chartColorViolet},
				},
			},
			{Name: "Services", Value: chartNodeService, Color: chartColorTeal},
			{Name: "Support", Value: chartNodeSupport, Color: chartColorAmber},
		},
	}
}

func nivoFlow(title string) nivo.FlowOptions {
	return nivo.FlowOptions{
		Title:       title,
		Description: "Source through middle to sink.",
		Nodes: []nivo.Node{
			{Name: "Source", Value: chartNodeProduct, Color: chartColorBlue},
			{Name: "Middle", Value: chartLinkIn, Color: chartColorTeal},
			{Name: "Sink", Value: chartLinkOut, Color: chartColorAmber},
		},
		Links: []nivo.Link{
			{Source: "Source", Target: "Middle", Value: chartLinkIn},
			{Source: "Middle", Target: "Sink", Value: chartLinkOut},
		},
	}
}

func nivoDays(title string) nivo.CalendarOptions {
	return nivo.CalendarOptions{
		Title:       title,
		Description: "Monday through Wednesday.",
		Days: []nivo.Day{
			{Label: "Mon", Value: chartDayMon},
			{Label: "Tue", Value: chartDayTue},
			{Label: "Wed", Value: chartDayWed},
		},
	}
}

func (a *builderApp) nivoPanel() *ui.Element {
	return a.chartGallery("Nivo", "Bar, line, pie, calendar, hierarchy, and flow charts.",
		widgetShell("Bar", chartSeriesBlurb, nivo.Bar(nivoSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, nivo.Line(nivoSeries("Line"))),
		widgetShell("Pie", chartSliceBlurb, nivo.Pie(nivoParts("Pie"))),
		widgetShell("Radar", chartSeriesBlurb, nivo.Radar(nivoSeries("Radar"))),
		widgetShell("Scatter Plot", chartSeriesBlurb, nivo.ScatterPlot(nivoSeries("Scatter Plot"))),
		widgetShell("Heat Map", chartSeriesBlurb, nivo.HeatMap(nivoSeries("Heat Map"))),
		widgetShell("Funnel", chartSliceBlurb, nivo.Funnel(nivoParts("Funnel"))),
		widgetShell("Calendar", chartDayBlurb, nivo.Calendar(nivoDays("Calendar"))),
		widgetShell("Bullet", chartSeriesBlurb, nivo.Bullet(nivoSeries("Bullet"))),
		widgetShell("Bump", chartSeriesBlurb, nivo.Bump(nivoSeries("Bump"))),
		widgetShell("Area Bump", chartSeriesBlurb, nivo.AreaBump(nivoSeries("Area Bump"))),
		widgetShell("Box Plot", chartSeriesBlurb, nivo.BoxPlot(nivoSeries("Box Plot"))),
		widgetShell("Choropleth", chartSliceBlurb, nivo.Choropleth(nivoParts("Choropleth"))),
		widgetShell("Circle Packing", chartTreeBlurb, nivo.CirclePacking(nivoTree("Circle Packing"))),
		widgetShell("Chord", chartFlowBlurb, nivo.Chord(nivoFlow("Chord"))),
		widgetShell("Geo Map", chartSliceBlurb, nivo.GeoMap(nivoParts("Geo Map"))),
		widgetShell("Icicle", chartTreeBlurb, nivo.Icicle(nivoTree("Icicle"))),
		widgetShell("Marimekko", chartSeriesBlurb, nivo.Marimekko(nivoSeries("Marimekko"))),
		widgetShell("Network", chartFlowBlurb, nivo.Network(nivoFlow("Network"))),
		widgetShell("Parallel Coordinates", chartSeriesBlurb, nivo.ParallelCoordinates(nivoSeries("Parallel Coordinates"))),
		widgetShell("Polar Bar", chartSeriesBlurb, nivo.PolarBar(nivoSeries("Polar Bar"))),
		widgetShell("Radial Bar", chartSliceBlurb, nivo.RadialBar(nivoParts("Radial Bar"))),
		widgetShell("Sankey", chartFlowBlurb, nivo.Sankey(nivoFlow("Sankey"))),
		widgetShell("Stream", chartSeriesBlurb, nivo.Stream(nivoSeries("Stream"))),
		widgetShell("Sunburst", chartTreeBlurb, nivo.Sunburst(nivoTree("Sunburst"))),
		widgetShell("Swarm Plot", chartSeriesBlurb, nivo.SwarmPlot(nivoSeries("Swarm Plot"))),
		widgetShell("Time Range", chartDayBlurb, nivo.TimeRange(nivoDays("Time Range"))),
		widgetShell("Tree", chartTreeBlurb, nivo.Tree(nivoTree("Tree"))),
		widgetShell("Tree Map", chartTreeBlurb, nivo.TreeMap(nivoTree("Tree Map"))),
		widgetShell("Voronoi", chartSeriesBlurb, nivo.Voronoi(nivoSeries("Voronoi"))),
		widgetShell("Waffle", chartSliceBlurb, nivo.Waffle(nivoParts("Waffle"))),
	)
}
