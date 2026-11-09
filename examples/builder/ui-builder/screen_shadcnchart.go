package main

import (
	"github.com/chinmay-sawant/ownframe/components/shadcnchart"
	"github.com/chinmay-sawant/ownframe/ui"
)

func shadcnSeries(title string) shadcnchart.Options {
	return shadcnchart.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []shadcnchart.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func shadcnParts(title string) shadcnchart.PartOptions {
	return shadcnchart.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []shadcnchart.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func (a *builderApp) shadcnchartPanel() *ui.Element {
	return a.chartGallery("shadcn charts", "Area, bar, line, pie, radar, and radial.",
		widgetShell("Area", chartSeriesBlurb, shadcnchart.Area(shadcnSeries("Area"))),
		widgetShell("Bar", chartSeriesBlurb, shadcnchart.Bar(shadcnSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, shadcnchart.Line(shadcnSeries("Line"))),
		widgetShell("Pie", chartSliceBlurb, shadcnchart.Pie(shadcnParts("Pie"))),
		widgetShell("Radar", chartSeriesBlurb, shadcnchart.Radar(shadcnSeries("Radar"))),
		widgetShell("Radial", chartSliceBlurb, shadcnchart.Radial(shadcnParts("Radial"))),
	)
}
