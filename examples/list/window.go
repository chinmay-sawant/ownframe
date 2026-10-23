// Package list computes a fixed-height row window for ownframe templates.
// ownframe has no virtual DOM: pass the window plus spacers to the template
// and call Page.SetWindowing and Page.SetScrollWindow.
package list

// Rows is a clamped window of list rows plus spacer heights in CSS pixels.
// Top is Start*rowH. Bottom is (totalRows-Start-Count)*rowH.
type Rows struct {
	Start  int
	Count  int
	Top    int
	Bottom int
}

// Window returns the rows to keep in the template for a scroll offset.
// overscan is extra rows above and below the visible band.
func Window(offsetY, viewH, gridY, rowH, totalRows, overscan int) Rows {
	if totalRows <= 0 || rowH <= 0 {
		return Rows{}
	}
	if overscan < 0 {
		overscan = 0
	}
	if viewH < 0 {
		viewH = 0
	}
	rel := offsetY - gridY
	if rel < 0 {
		rel = 0
	}
	start := rel/rowH - overscan
	if start < 0 {
		start = 0
	}
	// A floor of viewH/rowH drops a partial last row; +1 keeps it.
	// overscan applies once above the band and once below.
	count := viewH/rowH + overscan + overscan + 1
	if count > totalRows {
		count = totalRows
	}
	if start > totalRows-count {
		start = totalRows - count
	}
	return Rows{
		Start:  start,
		Count:  count,
		Top:    start * rowH,
		Bottom: (totalRows - start - count) * rowH,
	}
}
