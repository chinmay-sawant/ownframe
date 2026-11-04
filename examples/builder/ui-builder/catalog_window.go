package main

// catalog_window.go slices the ChadCN catalog into a visible row window so a
// scroll frame stops laying out all 77 cards.
//
// Precedent: examples/code-view owns the same trick via Page.SetScrollWindow
// plus Page.Scroll().SetWindowing(true). Its onScroll callback slices fixed
// 20px rows around the viewport (windowSliceForGroup in split_scroll.go) and
// a spacer preserves the total content height (SpaceHeight), so scrollbar
// metrics keep working while only visible rows become boxes.
//
// This catalog is a 2-column grid of variable-height cards, so the window is
// computed from a measured average row pitch instead of a fixed row height,
// with two overscan rows hiding estimation error. Geometry is pinned from
// full renders only: a.catalogGeo* fields hold the last full-render
// measurement, and windowed frames reuse it instead of re-measuring estimates
// off spacers (which would compound the pitch error). Tab switches and
// resizes clear the pin so the next full render re-pins. The design is
// otherwise stateless: the offset comes from a.center, the viewport from
// a.centerViewport, and the pitch/header from the pinned geometry. A full
// render measures; a windowed render carries spacers, and the content box
// keeps including them, so updateCenterScrollMetrics in scroll.go keeps
// reporting the full height with no change (spacers are sized so the box
// height is preserved up to pitch-estimation tolerance in the visible band;
// see catalogWindowSpacers and the plumbing test bound).

import (
	"github.com/chinmay-sawant/ownframe/ui"
)

const (
	// catalogWindowOverscanRows pads the visible row band above and below.
	// It absorbs pitch estimation error from variable card heights plus the
	// one-gap layout tolerance in the spacer math.
	catalogWindowOverscanRows = 2
	// catalogWindowGapPx mirrors the Gap(12) of the rows container in
	// catalogPanel; catalogWindowSectionGapPx mirrors the Gap(14) between
	// the heading block and the rows container.
	catalogWindowGapPx        = 12
	catalogWindowSectionGapPx = 14
	// catalogWindowTopSpacerID and catalogWindowBottomSpacerID mark the
	// height-preserving stubs. Their presence in Boxes tells geometry that
	// the last frame was a windowed render.
	catalogWindowTopSpacerID    = "catalog-window-top"
	catalogWindowBottomSpacerID = "catalog-window-bottom"
	// catalogWindowStubSlots reserves capacity for the two spacer stubs in
	// the windowed rows slice.
	catalogWindowStubSlots = 2
)

// catalogWindowPlan is one derived row window. windowed is false when the
// catalog is not measurable (other tab, no layout yet, degenerate boxes),
// and callers must render the full panel, which is today's behavior.
type catalogWindowPlan struct {
	startRow  int
	endRow    int // exclusive
	pitchPx   int
	rowsH     int
	headerPx  int
	totalRows int
	windowed  bool
}

// catalogRowCount maps an entry count to 2-column grid rows.
func catalogRowCount(totalEntries int) int {
	if totalEntries <= 0 {
		return 0
	}
	return (totalEntries + catalogColumns - 1) / catalogColumns
}

// catalogWindowRange slices totalRows into the [start, end) band covering the
// viewport at offsetY, padded by overscanRows on both sides. headerPx is the
// pixel offset where the rows start inside the scrolled content (the heading
// block above them), pitchPx the average row pitch including its gap.
// Degenerate inputs return the full range, matching today's render-all path.
func catalogWindowRange(totalRows, offsetY, viewH, headerPx, pitchPx, overscanRows int) (start, end int) {
	if totalRows <= 0 {
		return 0, 0
	}
	if pitchPx <= 0 || viewH <= 0 {
		return 0, totalRows
	}
	if overscanRows < 0 {
		overscanRows = 0
	}
	first := (offsetY-headerPx)/pitchPx - overscanRows
	if first < 0 {
		first = 0
	}
	if first > totalRows {
		first = totalRows
	}
	last := (offsetY+viewH-headerPx+pitchPx-1)/pitchPx + overscanRows
	if last < first {
		last = first
	}
	if last > totalRows {
		last = totalRows
	}
	return first, last
}

// catalogWindowEntries maps a [startRow, endRow) row band to the
// [startEntry, endEntry) catalog entries it covers.
func catalogWindowEntries(totalEntries, startRow, endRow int) (startEntry, endEntry int) {
	if totalEntries <= 0 || startRow < 0 || endRow < startRow {
		return 0, 0
	}
	startEntry = startRow * catalogColumns
	if startEntry > totalEntries {
		startEntry = totalEntries
	}
	endEntry = endRow * catalogColumns
	if endEntry > totalEntries {
		endEntry = totalEntries
	}
	return startEntry, endEntry
}

// catalogWindowSpacers sizes the height stubs above and below the visible
// rows so the rows container keeps its full height rowsH. pitchPx is the
// average row pitch including one gap; gapPx is the container gap. rowsH <= 0
// falls back to the uniform estimate. topPx + bottomPx + the visible block
// estimate ((endRow-startRow)*pitchPx - gapPx) equals rowsH exactly; the
// laid-out container additionally pays one gap per rendered spacer (at most
// two), which the overscan band and the content-box max in
// updateCenterScrollMetrics absorb. Zero-height edge stubs are reported as 0
// so the caller can omit them.
func catalogWindowSpacers(startRow, endRow, totalRows, pitchPx, gapPx, rowsH int) (topPx, bottomPx int) {
	if totalRows <= 0 || endRow <= startRow {
		return 0, 0
	}
	if rowsH <= 0 {
		rowsH = totalRows*pitchPx - gapPx
	}
	if rowsH < 0 {
		rowsH = 0
	}
	topPx = startRow * pitchPx
	if topPx < 0 {
		topPx = 0
	}
	if topPx > rowsH {
		topPx = rowsH
	}
	bottomPx = rowsH - endRow*pitchPx + gapPx
	if bottomPx < 0 {
		bottomPx = 0
	}
	return topPx, bottomPx
}

// catalogWindowGeometry measures the catalog layout from the last frame's
// boxes: headerPx is where the rows start inside the scrolled content,
// pitchPx the average row pitch, rowsH the full rows-container height.
//
// Only full renders are trusted: a windowed frame carries estimated spacers,
// so measuring pitch off it would compound the estimation error. The pinned
// full-render geometry is reused for windowed frames; a resize or tab switch
// produces a fresh full render and re-pins. ok is false when the catalog is
// not laid out and nothing is pinned yet, so callers fall back to the full
// render.
func (a *builderApp) catalogWindowGeometry() (headerPx, pitchPx, rowsH int, ok bool) {
	totalRows := catalogRowCount(len(catalogEntries()))
	if a.tab != "catalog" || totalRows == 0 {
		return 0, 0, 0, false
	}
	windowed := false
	var contentH float64
	var headingBottom, contentTop float64
	foundContent, foundHeading := false, false
	for _, box := range a.page.Boxes() {
		switch box.ID {
		case builderScrollContentID:
			contentH, contentTop = box.H, box.Y
			foundContent = true
		case builderActivePanelPrefix + "catalog":
			headingBottom = box.Y + box.H
			foundHeading = true
		case catalogWindowTopSpacerID, catalogWindowBottomSpacerID:
			windowed = true
		}
	}
	if windowed {
		if a.catalogGeoMeasured {
			return a.catalogGeoHeaderPx, a.catalogGeoPitchPx, a.catalogGeoRowsH, true
		}
		return 0, 0, 0, false
	}
	if !foundContent || !foundHeading {
		return 0, 0, 0, false
	}
	headerPx = int(headingBottom-contentTop) + catalogWindowSectionGapPx
	rowsH = int(contentH) - headerPx
	if rowsH <= 0 {
		return 0, 0, 0, false
	}
	pitchPx = (rowsH + catalogWindowGapPx) / totalRows
	if pitchPx <= 0 {
		return 0, 0, 0, false
	}
	a.catalogGeoMeasured = true
	a.catalogGeoHeaderPx, a.catalogGeoPitchPx, a.catalogGeoRowsH = headerPx, pitchPx, rowsH
	return headerPx, pitchPx, rowsH, true
}

// catalogWindowPlan derives the current row window from live state. The
// offset comes from the center ScrollState, the viewport from the laid-out
// boxes, and the pitch from the last frame's measurement.
func (a *builderApp) catalogWindowPlan() catalogWindowPlan {
	totalRows := catalogRowCount(len(catalogEntries()))
	plan := catalogWindowPlan{endRow: totalRows, totalRows: totalRows}
	headerPx, pitchPx, rowsH, ok := a.catalogWindowGeometry()
	if !ok {
		return plan
	}
	viewport := a.centerViewport()
	start, end := catalogWindowRange(totalRows, a.center.Offset(), int(viewport.H), headerPx, pitchPx, catalogWindowOverscanRows)
	if start <= 0 && end >= totalRows {
		plan.pitchPx, plan.rowsH, plan.headerPx = pitchPx, rowsH, headerPx
		return plan
	}
	plan.startRow, plan.endRow = start, end
	plan.pitchPx, plan.rowsH, plan.headerPx = pitchPx, rowsH, headerPx
	plan.windowed = true
	return plan
}

// catalogWindowSpacer builds one height-preserving stub. A non-positive
// height returns nil so the caller omits the edge stub.
func catalogWindowSpacer(id string, heightPx int) *ui.Element {
	if heightPx <= 0 {
		return nil
	}
	return ui.Div().ID(id).Height(ui.Px(float64(heightPx)))
}

// catalogWindowedRows builds only the visible card rows plus spacer stubs,
// ready to drop into the rows container in catalogPanel (see the TODO at the
// top of this file). When the layout is not measurable it builds every row,
// which is today's behavior. It calls catalogEntries read-only and reuses
// catalogCard and widgetsRow without modifying them.
func (a *builderApp) catalogWindowedRows() []*ui.Element {
	entries := catalogEntries()
	plan := a.catalogWindowPlan()
	startEntry, endEntry := catalogWindowEntries(len(entries), plan.startRow, plan.endRow)
	rows := make([]*ui.Element, 0, plan.endRow-plan.startRow+catalogWindowStubSlots)
	if !plan.windowed {
		startEntry, endEntry = 0, len(entries)
	}
	if topPx, bottomPx := catalogWindowSpacers(plan.startRow, plan.endRow, plan.totalRows, plan.pitchPx, catalogWindowGapPx, plan.rowsH); plan.windowed {
		if top := catalogWindowSpacer(catalogWindowTopSpacerID, topPx); top != nil {
			rows = append(rows, top)
		}
		for i := startEntry; i < endEntry; i += catalogColumns {
			cards := []*ui.Element{catalogCard(a, entries[i])}
			if i+1 < endEntry {
				cards = append(cards, catalogCard(a, entries[i+1]))
			}
			rows = append(rows, widgetsRow(cards...))
		}
		if bottom := catalogWindowSpacer(catalogWindowBottomSpacerID, bottomPx); bottom != nil {
			rows = append(rows, bottom)
		}
		return rows
	}
	for i := 0; i < len(entries); i += catalogColumns {
		cards := []*ui.Element{catalogCard(a, entries[i])}
		if i+1 < len(entries) {
			cards = append(cards, catalogCard(a, entries[i+1]))
		}
		rows = append(rows, widgetsRow(cards...))
	}
	return rows
}
