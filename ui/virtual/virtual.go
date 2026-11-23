// Package virtual computes visible row slices and spacer heights for fixed-row
// virtualized lists rendered with ownframe templates and SetScrollWindow.
package virtual

import "math"

// Range describes the visible half-open row slice [Start, End) and the spacer
// heights needed to preserve total scrollable content height.
type Range struct {
	Start        int
	End          int
	TopSpacer    float64
	BottomSpacer float64
	TotalHeight  float64
}

// Window computes the visible row range for a fixed-height list.
//
// scrollY is the current vertical scroll offset in CSS pixels, viewportH is
// the visible container height in CSS pixels, rowH is the height of one row in
// CSS pixels, totalRows is the full item count, and overscan is the number of
// extra rows to render above and below the viewport.
func Window(scrollY, viewportH, rowH float64, totalRows, overscan int) Range {
	if totalRows <= 0 || !(rowH > 0) || math.IsInf(rowH, 0) {
		return Range{}
	}
	if overscan < 0 {
		overscan = 0
	}
	// NaN fails every comparison, so reject it explicitly before clamping.
	if math.IsNaN(viewportH) || viewportH < 0 {
		viewportH = 0
	}
	if math.IsNaN(scrollY) {
		scrollY = 0
	}

	totalH := float64(totalRows) * rowH
	if viewportH > totalH {
		viewportH = totalH
	}
	maxScroll := math.Max(0, totalH-viewportH)
	if scrollY < 0 {
		scrollY = 0
	} else if scrollY > maxScroll {
		scrollY = maxScroll
	}

	// Clamp as floats first: a tiny rowH makes the quotient exceed int.
	rows := float64(totalRows)
	firstF := math.Min(math.Max(math.Floor(scrollY/rowH), 0), rows)
	first := int(firstF)
	visible := 0
	if viewportH > 0 {
		lastF := math.Min(math.Max(math.Ceil((scrollY+viewportH)/rowH), 0), rows)
		visible = int(lastF) - first
	}

	start := first - overscan
	if start < 0 {
		start = 0
	}
	end := first + visible + overscan
	if end > totalRows {
		end = totalRows
	}
	if end < start {
		end = start
	}

	return Range{
		Start:        start,
		End:          end,
		TopSpacer:    float64(start) * rowH,
		BottomSpacer: float64(totalRows-end) * rowH,
		TotalHeight:  totalH,
	}
}
