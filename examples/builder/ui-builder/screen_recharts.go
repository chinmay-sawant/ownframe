package main

import (
	"github.com/chinmay-sawant/ownframe/components/recharts"
	"github.com/chinmay-sawant/ownframe/ui"
)

func rechartsSeries(title string) recharts.Options {
	return recharts.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []recharts.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func rechartsParts(title string) recharts.PartOptions {
	return recharts.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []recharts.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func rechartsTree(title string) recharts.TreeOptions {
	return recharts.TreeOptions{
		Title:       title,
		Description: "Product, services, and support.",
		Nodes: []recharts.Node{
			{
				Name: "Product", Value: chartNodeProduct, Color: chartColorBlue,
				Children: []recharts.Node{
					{Name: "Core", Value: chartNodeCore, Color: chartColorBlue},
					{Name: "Add-ons", Value: chartNodeAddOn, Color: chartColorViolet},
				},
			},
			{Name: "Services", Value: chartNodeService, Color: chartColorTeal},
			{Name: "Support", Value: chartNodeSupport, Color: chartColorAmber},
		},
	}
}

func rechartsFlow(title string) recharts.FlowOptions {
	return recharts.FlowOptions{
		Title:       title,
		Description: "Source through middle to sink.",
		Nodes: []recharts.Node{
			{Name: "Source", Value: chartNodeProduct, Color: chartColorBlue},
			{Name: "Middle", Value: chartLinkIn, Color: chartColorTeal},
			{Name: "Sink", Value: chartLinkOut, Color: chartColorAmber},
		},
		Links: []recharts.Link{
			{Source: "Source", Target: "Middle", Value: chartLinkIn},
			{Source: "Middle", Target: "Sink", Value: chartLinkOut},
		},
	}
}

func (a *builderApp) rechartsPanel() *ui.Element {
	return a.chartGallery("Recharts", "Area, bar, line, composed, pie, radar, radial bar, "+
		"scatter, funnel, sankey, sunburst, and treemap.",
		widgetShell("Area", chartSeriesBlurb, recharts.Area(rechartsSeries("Area"))),
		widgetShell("Bar", chartSeriesBlurb, recharts.Bar(rechartsSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, recharts.Line(rechartsSeries("Line"))),
		widgetShell("Composed", chartSeriesBlurb, recharts.Composed(rechartsSeries("Composed"))),
		widgetShell("Pie", chartSliceBlurb, recharts.Pie(rechartsParts("Pie"))),
		widgetShell("Radar", chartSeriesBlurb, recharts.Radar(rechartsSeries("Radar"))),
		widgetShell("Radial Bar", chartSliceBlurb, recharts.RadialBar(rechartsParts("Radial Bar"))),
		widgetShell("Scatter", chartSeriesBlurb, recharts.Scatter(rechartsSeries("Scatter"))),
		widgetShell("Funnel", chartSliceBlurb, recharts.Funnel(rechartsParts("Funnel"))),
		widgetShell("Sankey", chartFlowBlurb, recharts.Sankey(rechartsFlow("Sankey"))),
		widgetShell("Sunburst", chartTreeBlurb, recharts.Sunburst(rechartsTree("Sunburst"))),
		widgetShell("Treemap", chartTreeBlurb, recharts.Treemap(rechartsTree("Treemap"))),
	)
}
