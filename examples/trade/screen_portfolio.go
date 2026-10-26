package main

import (
	"strconv"
	"strings"
)

// The performance line is authored in a 1240 x 260 design box. The CSS stage
// keeps that ratio, so a coordinate becomes a percentage of the design width
// or height and stays on the same vertex at every window size.
const (
	pfChartDesignWidth  = 1240.0
	pfChartDesignHeight = 260.0
	// pfChartMidY is the middle of the design box, used when no point parses.
	pfChartMidY    = pfChartDesignHeight / 2
	pfChartPercent = 100.0
	pfChartDigits  = 3
	// pfLineHalf is half of a line's weight. A segment is lifted by this much
	// so its centerline lands on the plotted vertex.
	pfLineHalf = 0.5
)

// PFSegmentX places one performance segment's start vertex across the stage.
func (v View) PFSegmentX(seg ChartSegment) string {
	return pfNumber(pfPercent(seg.Left, pfChartDesignWidth))
}

// PFSegmentY places one performance segment's start vertex down the stage.
func (v View) PFSegmentY(seg ChartSegment) string {
	return pfNumber(pfPercent(seg.Top, pfChartDesignHeight))
}

// PFSegmentWidth is the segment length as a share of the design width, so a
// span always ends on the next plotted point.
func (v View) PFSegmentWidth(seg ChartSegment) string {
	return pfNumber(pfPercent(seg.Length, pfChartDesignWidth))
}

// PFSegmentTop is half the line weight, in pixels. Lifting the line by half
// puts its centerline on the start vertex instead of its top edge.
func (v View) PFSegmentTop(seg ChartSegment) string {
	return pfNumber(-seg.Width * pfLineHalf)
}

// PFGuideY places one horizontal grid line and its y-axis label. The guide
// list uses the same design height as the plotted segments.
func (v View) PFGuideY(index int) string {
	grid := v.Portfolio.Performance.Grid
	if index < 0 || index >= len(grid) {
		return "0"
	}

	return pfNumber(pfPercent(grid[index].Y1, pfChartDesignHeight))
}

// PFLastPointX returns the final plotted x as a share of the design width.
func (v View) PFLastPointX() string {
	x, _ := pfLastPoint(v.Portfolio.Performance.Points)
	if x < 0 || x > pfChartDesignWidth {
		x = pfChartDesignWidth
	}

	return pfNumber(pfPercent(x, pfChartDesignWidth))
}

// PFLastPointY returns the final plotted y as a share of the design height.
func (v View) PFLastPointY() string {
	_, y := pfLastPoint(v.Portfolio.Performance.Points)
	if y < 0 || y > pfChartDesignHeight {
		y = pfChartMidY
	}

	return pfNumber(pfPercent(y, pfChartDesignHeight))
}

// pfLastPoint reads the final x,y pair from the performance point string.
func pfLastPoint(points string) (float64, float64) {
	fields := strings.Fields(points)
	if len(fields) == 0 {
		return pfChartDesignWidth, pfChartMidY
	}

	left, right, ok := strings.Cut(fields[len(fields)-1], ",")
	if !ok {
		return pfChartDesignWidth, pfChartMidY
	}

	x, errX := strconv.ParseFloat(strings.TrimSpace(left), 64)
	y, errY := strconv.ParseFloat(strings.TrimSpace(right), 64)
	if errX != nil || errY != nil {
		return pfChartDesignWidth, pfChartMidY
	}

	return x, y
}

// pfPercent converts a design coordinate into a percentage of its axis.
func pfPercent(value, total float64) float64 {
	if total == 0 {
		return 0
	}

	return value / total * pfChartPercent
}

// pfNumber formats one CSS number with enough precision for the longest
// stage the portfolio supports.
func pfNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', pfChartDigits, 64)
}
