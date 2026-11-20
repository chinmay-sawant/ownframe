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
	if totalRows <= 0 || rowH <= 0 {
		return Range{}
	}
	if overscan < 0 {
		overscan = 0
	}
	if viewportH < 0 {
		viewportH = 0
	}

	totalH := float64(totalRows) * rowH
	maxScroll := math.Max(0, totalH-viewportH)
	if scrollY < 0 {
		scrollY = 0
	} else if scrollY > maxScroll {
		scrollY = maxScroll
	}

	first := int(math.Floor(scrollY / rowH))
	visible := 0
	if viewportH > 0 {
		visible = int(math.Ceil((scrollY+viewportH)/rowH)) - first
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
