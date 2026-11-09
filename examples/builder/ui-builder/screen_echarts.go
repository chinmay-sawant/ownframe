package main

import (
	"github.com/chinmay-sawant/ownframe/components/echarts"
	"github.com/chinmay-sawant/ownframe/ui"
)

func echartsSeries(title string) echarts.Options {
	return echarts.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []echarts.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func echartsParts(title string) echarts.PartOptions {
	return echarts.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []echarts.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func echartsTree(title string) echarts.TreeOptions {
	return echarts.TreeOptions{
		Title:       title,
		Description: "Product, services, and support.",
		Nodes: []echarts.Node{
			{
				Name: "Product", Value: chartNodeProduct, Color: chartColorBlue,
				Children: []echarts.Node{
					{Name: "Core", Value: chartNodeCore, Color: chartColorBlue},
					{Name: "Add-ons", Value: chartNodeAddOn, Color: chartColorViolet},
				},
			},
			{Name: "Services", Value: chartNodeService, Color: chartColorTeal},
			{Name: "Support", Value: chartNodeSupport, Color: chartColorAmber},
		},
	}
}

func echartsFlow(title string) echarts.FlowOptions {
	return echarts.FlowOptions{
		Title:       title,
		Description: "Source through middle to sink.",
		Nodes: []echarts.Node{
			{Name: "Source", Value: chartNodeProduct, Color: chartColorBlue},
			{Name: "Middle", Value: chartLinkIn, Color: chartColorTeal},
			{Name: "Sink", Value: chartLinkOut, Color: chartColorAmber},
		},
		Links: []echarts.Link{
			{Source: "Source", Target: "Middle", Value: chartLinkIn},
			{Source: "Middle", Target: "Sink", Value: chartLinkOut},
		},
	}
}

func echartsCandles(title string) echarts.CandleOptions {
	return echarts.CandleOptions{
		Title:       title,
		Description: "Three sessions.",
		Points: []echarts.OHLC{
			{Label: "Mon", Open: chartOpenMon, High: chartHighMon, Low: chartLowMon, Close: chartCloseMon},
			{Label: "Tue", Open: chartOpenTue, High: chartHighTue, Low: chartLowTue, Close: chartCloseTue},
			{Label: "Wed", Open: chartOpenWed, High: chartHighWed, Low: chartLowWed, Close: chartCloseWed},
		},
	}
}

func (a *builderApp) echartsPanel() *ui.Element {
	return a.chartGallery("ECharts", "Cartesian, statistical, map, tree, graph, and gauge charts.",
		widgetShell("Bar", chartSeriesBlurb, echarts.Bar(echartsSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, echarts.Line(echartsSeries("Line"))),
		widgetShell("Scatter", chartSeriesBlurb, echarts.Scatter(echartsSeries("Scatter"))),
		widgetShell("Pie", chartSliceBlurb, echarts.Pie(echartsParts("Pie"))),
		widgetShell("Radar", chartSeriesBlurb, echarts.Radar(echartsSeries("Radar"))),
		widgetShell("Heatmap", chartSeriesBlurb, echarts.Heatmap(echartsSeries("Heatmap"))),
		widgetShell("Candlestick", chartCandleBlurb, echarts.Candlestick(echartsCandles("Candlestick"))),
		widgetShell("Boxplot", chartSeriesBlurb, echarts.Boxplot(echartsSeries("Boxplot"))),
		widgetShell("Gauge", chartValueBlurb, echarts.Gauge("Load", chartGauge)),
		widgetShell("Funnel", chartSliceBlurb, echarts.Funnel(echartsParts("Funnel"))),
		widgetShell("Effect Scatter", chartSeriesBlurb, echarts.EffectScatter(echartsSeries("Effect Scatter"))),
		widgetShell("Pictorial Bar", chartSeriesBlurb, echarts.PictorialBar(echartsSeries("Pictorial Bar"))),
		widgetShell("Map", chartSliceBlurb, echarts.Map(echartsParts("Map"))),
		widgetShell("Tree", chartTreeBlurb, echarts.Tree(echartsTree("Tree"))),
		widgetShell("Treemap", chartTreeBlurb, echarts.Treemap(echartsTree("Treemap"))),
		widgetShell("Graph", chartFlowBlurb, echarts.Graph(echartsFlow("Graph"))),
		widgetShell("Chord", chartFlowBlurb, echarts.Chord(echartsFlow("Chord"))),
		widgetShell("Parallel", chartSeriesBlurb, echarts.Parallel(echartsSeries("Parallel"))),
		widgetShell("Sankey", chartFlowBlurb, echarts.Sankey(echartsFlow("Sankey"))),
		widgetShell("Lines", chartFlowBlurb, echarts.Lines(echartsFlow("Lines"))),
		widgetShell("Theme River", chartSeriesBlurb, echarts.ThemeRiver(echartsSeries("Theme River"))),
		widgetShell("Sunburst", chartTreeBlurb, echarts.Sunburst(echartsTree("Sunburst"))),
		widgetShell("Custom", chartSeriesBlurb, echarts.Custom(echartsSeries("Custom"))),
	)
}
