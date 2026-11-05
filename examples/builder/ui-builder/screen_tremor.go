package main

import (
	"github.com/chinmay-sawant/ownframe/components/tremor"
	"github.com/chinmay-sawant/ownframe/ui"
)

func tremorSeries(title string) tremor.Options {
	return tremor.Options{
		Title:       title,
		Description: "Opened and closed, Monday through Wednesday.",
		Labels:      []string{"Mon", "Tue", "Wed"},
		Series: []tremor.Series{
			{Name: "Opened", Values: []float64{chartMon, chartTue, chartWed}, Color: chartColorBlue},
			{Name: "Closed", Values: []float64{chartMonB, chartTueB, chartWedB}, Color: chartColorTeal},
		},
	}
}

func tremorParts(title string) tremor.PartOptions {
	return tremor.PartOptions{
		Title:       title,
		Description: "Desktop, mobile, and tablet share.",
		Slices: []tremor.Slice{
			{Name: "Desktop", Value: chartDesktop, Color: chartColorBlue},
			{Name: "Mobile", Value: chartMobile, Color: chartColorTeal},
			{Name: "Tablet", Value: chartTablet, Color: chartColorAmber},
		},
	}
}

func tremorTracker(title string) tremor.TrackerOptions {
	return tremor.TrackerOptions{
		Title:       title,
		Description: "Three good days.",
		Cells: []tremor.TrackerCell{
			{Name: "Mon", Status: "good"},
			{Name: "Tue", Status: "good"},
			{Name: "Wed", Status: "good"},
		},
	}
}

func (a *builderApp) tremorPanel() *ui.Element {
	return a.chartGallery("Tremor", "Area, bar, line, combo, donut, sparks, lists, tracker, and progress.",
		widgetShell("Area", chartSeriesBlurb, tremor.Area(tremorSeries("Area"))),
		widgetShell("Bar", chartSeriesBlurb, tremor.Bar(tremorSeries("Bar"))),
		widgetShell("Line", chartSeriesBlurb, tremor.Line(tremorSeries("Line"))),
		widgetShell("Combo", chartSeriesBlurb, tremor.Combo(tremorSeries("Combo"))),
		widgetShell("Donut", chartSliceBlurb, tremor.Donut(tremorParts("Donut"))),
		widgetShell("Spark Area", chartSeriesBlurb, tremor.SparkArea(tremorSeries("Spark Area"))),
		widgetShell("Spark Line", chartSeriesBlurb, tremor.SparkLine(tremorSeries("Spark Line"))),
		widgetShell("Spark Bar", chartSeriesBlurb, tremor.SparkBar(tremorSeries("Spark Bar"))),
		widgetShell("Bar List", chartSliceBlurb, tremor.BarList(tremorParts("Bar List"))),
		widgetShell("Category Bar", chartSliceBlurb, tremor.CategoryBar(tremorParts("Category Bar"))),
		widgetShell("Tracker", chartTrackBlurb, tremor.Tracker(tremorTracker("Tracker"))),
		widgetShell("Progress Bar", chartValueBlurb, tremor.ProgressBar("Uptime", chartGauge)),
		widgetShell("Progress Circle", chartValueBlurb, tremor.ProgressCircle("Capacity", chartGauge)),
	)
}
